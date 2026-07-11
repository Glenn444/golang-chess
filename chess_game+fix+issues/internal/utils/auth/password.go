package auth


import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string,error){
	passwordByte, err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err != nil{
		return "",err
	}
	return string(passwordByte), nil
}

func CheckPassword(HashedPassword string, password string)error{
	return bcrypt.CompareHashAndPassword([]byte(HashedPassword),[]byte(password))
}