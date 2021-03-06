package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDBRestore(t *testing.T) {
	tmpDir := path.Join(os.TempDir(), "neogo.restoretest")
	require.NoError(t, os.Mkdir(tmpDir, os.ModePerm))
	defer os.RemoveAll(tmpDir)

	chainPath := path.Join(tmpDir, "neogotestchain")
	cfg, err := config.LoadFile("../config/protocol.unit_testnet.yml")
	require.NoError(t, err, "could not load config")
	cfg.ApplicationConfiguration.DBConfiguration.Type = "leveldb"
	cfg.ApplicationConfiguration.DBConfiguration.LevelDBOptions.DataDirectoryPath = chainPath

	out, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	cfgPath := path.Join(tmpDir, "protocol.unit_testnet.yml")
	require.NoError(t, ioutil.WriteFile(cfgPath, out, os.ModePerm))

	// generated via `go run ./scripts/gendump/main.go --out ./cli/testdata/chain50x2.acc --blocks 50 --txs 2`
	const inDump = "./testdata/chain50x2.acc"
	e := newExecutor(t, false)
	defer e.Close(t)
	stateDump := path.Join(tmpDir, "neogo.teststate")
	baseArgs := []string{"neo-go", "db", "restore", "--unittest",
		"--config-path", tmpDir, "--in", inDump, "--dump", stateDump}

	// First 15 blocks.
	e.Run(t, append(baseArgs, "--count", "15")...)

	// Invalid skips.
	e.RunWithError(t, append(baseArgs, "--count", "10", "--skip", "14")...)
	e.RunWithError(t, append(baseArgs, "--count", "10", "--skip", "16")...)
	// Big count.
	e.RunWithError(t, append(baseArgs, "--count", "1000", "--skip", "15")...)

	// Continue 15..25
	e.Run(t, append(baseArgs, "--count", "10", "--skip", "15")...)

	// Continue till end.
	e.Run(t, append(baseArgs, "--skip", "25")...)

	// Dump and compare.
	dumpPath := path.Join(tmpDir, "testdump.acc")
	e.Run(t, "neo-go", "db", "dump", "--unittest",
		"--config-path", tmpDir, "--out", dumpPath)

	d1, err := ioutil.ReadFile(inDump)
	require.NoError(t, err)
	d2, err := ioutil.ReadFile(dumpPath)
	require.NoError(t, err)
	require.Equal(t, d1, d2, "dumps differ")
}
