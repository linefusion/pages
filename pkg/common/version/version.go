package version

var (
	appVersion = "dev"
	appCommit  = ""
	appDate    = ""
	appBuiltBy = ""
)

func Initialize(version, commit, date, builtBy string) {
	appVersion = version
	appCommit = commit
	appDate = date
	appBuiltBy = builtBy
}

func GetVersion() string {
	return appVersion
}

func GetCommit() string {
	return appCommit
}

func GetDate() string {
	return appDate
}

func GetBuiltBy() string {
	return appBuiltBy
}
