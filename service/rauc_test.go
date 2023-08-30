package service_test

// TODO 把这个修一下
// func TestVerifyRAUC(t *testing.T) {
// 	logger.LogInitConsoleOnly()

// 	tmpDir, err := os.MkdirTemp("", "casaos-verify-rauc-*")
// 	defer os.RemoveAll(tmpDir)
// 	assert.NoError(t, err)
// 	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
// 	assert.NoError(t, err)

// 	_, err = service.VerifyRAUC(*release)
// 	assert.ErrorContains(t, err, "not found")

// 	// TODO to download rauc

// 	// TODO to verify rauc again
// }
