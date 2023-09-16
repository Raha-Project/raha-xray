package services

import (
	"raha-xray/database"
	"raha-xray/database/model"

	"gorm.io/gorm"
)

type RuleService struct {
}

func (s *RuleService) GetAll() ([]*model.Rule, error) {
	db := database.GetDB()
	var rules []*model.Rule
	err := db.Model(model.Rule{}).Find(&rules).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return rules, nil
}

func (s *RuleService) Get(id int) (*model.Rule, error) {
	db := database.GetDB()
	var rule *model.Rule
	err := db.Model(model.Rule{}).Where("id = ?", id).Find(&rule).Error
	if err != nil {
		return nil, err
	}
	if rule.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return rule, nil
}

func (s *RuleService) Save(rule *model.Rule) error {
	db := database.GetDB()
	return db.Save(rule).Error
}

func (s *RuleService) Del(id uint) error {
	db := database.GetDB()
	return db.Delete(model.Rule{}, id).Error
}
