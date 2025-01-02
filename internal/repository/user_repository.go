package repository

import "github.com/rohan03122001/quizzing/internal/models"

type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(user *models.User) error{
	return r.db.Create(user).Error
}


// ENDGOAL WE WILL RETURN A USER AND ERROR OPTIONAL. 
//so we will need a vessle to store our query result then we find the id item and store it in the vessle
func(r *UserRepository) GetByID(id string) (*models.User, error){
	var user models.User

	if err:= r.db.First(&user, "id=?", id).Error; err!=nil{
		return nil,err
	}

	return &user, nil
}

func(r *UserRepository) GetByUsername(username string) (*models.User, error){
	var user models.User

	if err:= r.db.First(&user, "username=?", username).Error; err!=nil{
		return nil,err
	}

	return &user, nil
}



// delete user
func(r *UserRepository) DeleteUser(id string) error{
	return r.db.Delete(&models.User{},"id=?", id).Error
}

// Update modifies an existing user
func (r *UserRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

