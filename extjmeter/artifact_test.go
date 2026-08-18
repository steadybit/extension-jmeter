/*
 * Copyright 2026 steadybit GmbH. All rights reserved.
 */

package extjmeter

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_appendFileArtifact_skips_missing_file(t *testing.T) {
	artifacts, err := appendFileArtifact(nil, filepath.Join(t.TempDir(), "absent.txt"), "log.txt")

	require.NoError(t, err)
	assert.Empty(t, artifacts)
}

func Test_appendFileArtifact_attaches_small_file_as_is(t *testing.T) {
	path := writeFile(t, "log.txt", []byte("some log"))

	artifacts, err := appendFileArtifact(nil, path, "prefix_log.txt")

	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.Equal(t, "prefix_log.txt", artifacts[0].Label)
	assert.Equal(t, "some log", string(decode(t, artifacts[0])))
}

func Test_appendFileArtifact_zips_large_file(t *testing.T) {
	content := bytes.Repeat([]byte("a"), artifactZipThreshold+1)
	path := writeFile(t, "log.txt", content)

	artifacts, err := appendFileArtifact(nil, path, "prefix_log.txt")

	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.Equal(t, "prefix_log.zip", artifacts[0].Label)
	assert.FileExists(t, filepath.Join(filepath.Dir(path), "log.zip"))

	data := decode(t, artifacts[0])
	assert.Less(t, len(data), len(content), "the archive should be smaller than the file")

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	require.Len(t, r.File, 1)
	assert.Equal(t, "log.txt", r.File[0].Name, "the archive must not contain the file's directories")

	entry, err := r.File[0].Open()
	require.NoError(t, err)
	defer func() { _ = entry.Close() }()
	unzipped, err := io.ReadAll(entry)
	require.NoError(t, err)
	assert.Equal(t, content, unzipped)
}

func writeFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, content, 0644))
	return path
}

func decode(t *testing.T, artifact action_kit_api.Artifact) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(artifact.Data)
	require.NoError(t, err)
	return data
}
