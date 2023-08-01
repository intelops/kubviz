package model

type GithubImage struct {
	PackageId    string `json:"package_id"`
	CreatedAt    string `json:"created_at"`
	ImageName    string `json:"image_name"`
	Organisation string `json:"organisation"`
	UpdatedAt    string `json:"updated_at"`
	Visibility   string `json:"visibility"`
	ShaID        string `json:"sha_id"`
	ImageId      string `json:"image_id"`
}
