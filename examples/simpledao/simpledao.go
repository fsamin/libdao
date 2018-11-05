package simpledao

//MyType represents a row of table my_type
//libdao table=my_type
type MyType struct {
	ID  int64  `db:"id,primarykey,autoincrement"`
	Foo string `db:"foo"`
	Bar string `db:"bar"`
	Biz string `db:"biz,json"`
}

//MyTypeDAO is the dao for MyType
type MyTypeDAO interface {
	Insert(*MyType) error
	Update(*MyType) error
	Delete(*MyType) error
	All() ([]MyType, error)
	FindByID(int64) (*MyType, error)
	FindByFoo(string) ([]MyType, error)
	FindByFooAndBar(string, string) ([]MyType, error)
}
