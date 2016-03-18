package dmp

type Member struct {
	Members []string `json:"members"`
}

func CreateMember(members []string) *Member {
	return &Member{
		Members: members,
	}
}
