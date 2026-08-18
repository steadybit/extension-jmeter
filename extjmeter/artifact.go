/*
 * Copyright 2026 steadybit GmbH. All rights reserved.
 */

package extjmeter

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extfile"
)

// artifactZipThreshold is the size above which an artifact is zipped before being
// base64 encoded into the response.
const artifactZipThreshold = 1000000

// appendFileArtifact appends the file at path as an artifact named label. Files
// larger than artifactZipThreshold are zipped first, with ".zip" replacing the
// extension in both the file name and the label. A missing file is skipped.
func appendFileArtifact(artifacts []action_kit_api.Artifact, path, label string) ([]action_kit_api.Artifact, error) {
	stats, err := os.Stat(path)
	if err != nil {
		return artifacts, nil
	}

	if stats.Size() > artifactZipThreshold {
		zipped := replaceExtension(path, ".zip")
		log.Info().Msgf("Zipping %s to %s", path, zipped)
		if err := zipFile(path, zipped); err != nil {
			return artifacts, extension_kit.ToError(fmt.Sprintf("Failed to zip %s", path), err)
		}
		path, label = zipped, replaceExtension(label, ".zip")
	}

	content, err := extfile.File2Base64(path)
	if err != nil {
		return artifacts, err
	}
	return append(artifacts, action_kit_api.Artifact{Label: label, Data: content}), nil
}

func replaceExtension(path, ext string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}

// zipFile writes src into a newly created zip archive at dst, stored under src's
// base name.
func zipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	w := zip.NewWriter(out)
	entry, err := w.Create(filepath.Base(src))
	if err != nil {
		return err
	}
	if _, err := io.Copy(entry, in); err != nil {
		return err
	}
	return w.Close()
}
