package infra

type UserModel struct {
	ID       string `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}
