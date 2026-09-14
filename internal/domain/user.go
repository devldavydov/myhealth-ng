package domain

type UserIdentity struct {
	GUID                   string `json:"guid"`
	Name                   string `json:"name"`
	CertificateFingerprint string `json:"certificateFingerprint"`
	LastSeenAt             string `json:"lastSeenAt"`
}
