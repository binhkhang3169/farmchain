package models

import (
	"gorm.io/gorm"
)

type AgriculturalArea struct {
	gorm.Model
	CropType     string `json:"cropType"`
	PlantingDate string `json:"plantingDate"`
	HarvestDate  string `json:"harvestDate"`
	GeoJSON      string `json:"geoJson"` // Ta lưu polygon dưới dạng GeoJSON string
}
