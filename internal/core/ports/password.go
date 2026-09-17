package ports

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type PasswordVerifier interface {
	Verify(password, hashedPassword string) (bool, error)
}

type PasswordManager interface {
	PasswordHasher
	PasswordVerifier
}
