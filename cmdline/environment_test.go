package cmdline

type mockEnvironment struct{}

func (*mockEnvironment) Get(key string) string {
	if key == "HOME" {
		return mockHomeDir
	}
	return ""
}

func (*mockEnvironment) List() []string {
	return []string{"HOME=" + mockHomeDir}
}
