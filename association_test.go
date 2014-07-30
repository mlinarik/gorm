package gorm_test

import "testing"

func TestSubStruct(t *testing.T) {
	db.DropTable(Category{})
	db.DropTable(Post{})
	db.DropTable(Comment{})

	db.CreateTable(Category{})
	db.CreateTable(Post{})
	db.CreateTable(Comment{})

	post := Post{
		Title:        "post 1",
		Body:         "body 1",
		Comments:     []Comment{{Content: "Comment 1"}, {Content: "Comment 2"}},
		Category:     Category{Name: "Category 1"},
		MainCategory: Category{Name: "Main Category 1"},
	}

	if err := db.Save(&post).Error; err != nil {
		t.Errorf("Got errors when save post")
	}

	if db.First(&Category{}, "name = ?", "Category 1").Error != nil {
		t.Errorf("Category should be saved")
	}

	var p Post
	db.First(&p, post.Id)

	if post.CategoryId.Int64 == 0 || p.CategoryId.Int64 == 0 || post.MainCategoryId == 0 || p.MainCategoryId == 0 {
		t.Errorf("Category Id should exist")
	}

	if db.First(&Comment{}, "content = ?", "Comment 1").Error != nil {
		t.Errorf("Comment 1 should be saved")
	}
	if post.Comments[0].PostId == 0 {
		t.Errorf("Comment Should have post id")
	}

	var comment Comment
	if db.First(&comment, "content = ?", "Comment 2").Error != nil {
		t.Errorf("Comment 2 should be saved")
	}

	if comment.PostId == 0 {
		t.Errorf("Comment 2 Should have post id")
	}

	comment3 := Comment{Content: "Comment 3", Post: Post{Title: "Title 3", Body: "Body 3"}}
	db.Save(&comment3)
}

func TestRelated(t *testing.T) {
	user := User{
		Name:            "jinzhu",
		BillingAddress:  Address{Address1: "Billing Address - Address 1"},
		ShippingAddress: Address{Address1: "Shipping Address - Address 1"},
		Emails:          []Email{{Email: "jinzhu@example.com"}, {Email: "jinzhu-2@example@example.com"}},
		CreditCard:      CreditCard{Number: "1234567890"},
	}

	db.Save(&user)

	if user.CreditCard.Id == 0 {
		t.Errorf("After user save, credit card should have id")
	}

	if user.BillingAddress.Id == 0 {
		t.Errorf("After user save, billing address should have id")
	}

	if user.Emails[0].Id == 0 {
		t.Errorf("After user save, billing address should have id")
	}

	var emails []Email
	db.Model(&user).Related(&emails)
	if len(emails) != 2 {
		t.Errorf("Should have two emails")
	}

	var emails2 []Email
	db.Model(&user).Where("email = ?", "jinzhu@example.com").Related(&emails2)
	if len(emails2) != 1 {
		t.Errorf("Should have two emails")
	}

	var user1 User
	db.Model(&user).Related(&user1.Emails)
	if len(user1.Emails) != 2 {
		t.Errorf("Should have only one email match related condition")
	}

	var address1 Address
	db.Model(&user).Related(&address1, "BillingAddressId")
	if address1.Address1 != "Billing Address - Address 1" {
		t.Errorf("Should get billing address from user correctly")
	}

	user1 = User{}
	db.Model(&address1).Related(&user1, "BillingAddressId")
	if db.NewRecord(user1) {
		t.Errorf("Should get user from address correctly")
	}

	var user2 User
	db.Model(&emails[0]).Related(&user2)
	if user2.Id != user.Id || user2.Name != user.Name {
		t.Errorf("Should get user from email correctly")
	}

	var creditcard CreditCard
	var user3 User
	db.Debug().First(&creditcard, "number = ?", "1234567890")
	db.Debug().Model(&creditcard).Related(&user3)
	if user3.Id != user.Id || user3.Name != user.Name {
		t.Errorf("Should get user from credit card correctly")
	}

	if !db.Model(&CreditCard{}).Related(&User{}).RecordNotFound() {
		t.Errorf("RecordNotFound for Related")
	}
}

func TestQueryManyToManyWithRelated(t *testing.T) {
	var languages = []Language{{Name: "ZH"}, {Name: "EN"}, {Name: "DE"}}
	user := User{Name: "Many2Many", Languages: languages}
	db.Save(&user)

	var newLanguages []Language
	// db.Model(&user).Related(&newLanguages, "Languages")
	// if len(newLanguages) != 3 {
	// 	t.Errorf("Query many to many relations")
	// }

	newLanguages = []Language{}
	db.Model(&user).Association("Languages").Find(&newLanguages)
	if len(newLanguages) != 3 {
		t.Errorf("Query many to many relations")
	}

	// db.Model(&User{}).Many2Many("Languages").Add(&Language{})
	// db.Model(&User{}).Many2Many("Languages").Remove(&Language{})
	// db.Model(&User{}).Many2Many("Languages").Replace(&[]Language{})
	// db.Model(&User{}).Related(&[]Language{}, "Languages")
	// SELECT `languages`.* FROM `languages` INNER JOIN `user_languages` ON `languages`.`id` = `user_languages`.`language_id` WHERE `user_languages`.`user_id` = 111
}