package enum

// Country
//type CountryCode string

var CountryMaps = map[string]string{
	"nl": "Netherlands",
	"de": "German",
	"fr": "France",
	"es": "Spain",
	"ch": "Switzerland",
	"fi": "Finland",
	"se": "Sweden",
	"no": "Norway",
	"dk": "Denmark",
	"uk": "United Kingdom",
	"gb": "United Kingdom",
	"nz": "New Zealand",
	"au": "Australia",
	"ca": "Canada",
}

// Site

type Site string

func (s Site) String() string {
	return string(s)
}

const (
	SiteIndeed        Site = "indeed"
	SiteStackoverflow Site = "stackoverflow"
	SiteLinkedin      Site = "linkedin"
)

var Sites = []Site{
	"",
	SiteIndeed,
	SiteStackoverflow,
	SiteLinkedin,
}
