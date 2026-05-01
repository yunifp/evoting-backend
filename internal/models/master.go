package models


type RefWilayah struct {
	WilayahID    string `gorm:"column:wilayah_id;primaryKey"`
	Parent       string `gorm:"column:parent"`
	Children     string `gorm:"column:children"`
	NamaWilayah  string `gorm:"column:nama_wilayah"`
	UsulanNama   string `gorm:"column:usulan_nama"`
	Tingkat      int    `gorm:"column:tingkat"`
	TingkatLabel string `gorm:"column:tingkat_label"`
	KodePro      string `gorm:"column:kode_pro"`
	KodeKab      string `gorm:"column:kode_kab"`
	KodeKec      string `gorm:"column:kode_kec"`
	KodeKel      string `gorm:"column:kode_kel"`
	Singkatan    string `gorm:"column:singkatan"`
	Lat          string `gorm:"column:lat"`
	Lon          string `gorm:"column:lon"`
}

func (RefWilayah) TableName() string {
	return "ref_wilayah"
}

type StatusKawin struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
}