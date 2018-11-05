//generated file
package simpledao

import "github.com/fsamin/libdao"

type myTypeDAO struct {
	db    libdao.SqlExecutor
	dbCtx libdao.SqlExecutorWithContext
}

func (receiver *myTypeDAO) Insert(target *MyType) error {
	q := "INSERT INTO my_type (\"bar\",\"biz\",\"foo\") VALUE ($1,$2,$3) RETURNING id"
	if err := receiver.db.QueryRow(q, target.Bar, target.Biz, target.Foo).Scan(&target.ID); err != nil {
		return err
	}
	return nil
}
func (receiver *myTypeDAO) Update(*MyType) error {
	return nil
}
func (receiver *myTypeDAO) Delete(*MyType) error {
	return nil
}
func (receiver *myTypeDAO) All() ([]MyType, error) {
	return nil, nil
}
func (receiver *myTypeDAO) FindByID(int64) (*MyType, error) {
	return nil, nil
}
func (receiver *myTypeDAO) FindByFoo(string) ([]MyType, error) {
	return nil, nil
}
func (receiver *myTypeDAO) FindByFooAndBar(string, string) ([]MyType, error) {
	return nil, nil
}

func NewMyTypeDAO(db libdao.SqlExecutor) MyTypeDAO {
	return &myTypeDAO{db: db}
}

func NewMyTypeDAOWithContext(db libdao.SqlExecutorWithContext) MyTypeDAO {
	return &myTypeDAO{dbCtx: db}
}
