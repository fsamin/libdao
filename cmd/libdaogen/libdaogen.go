package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/fsamin/go-typeparser"
)

var (
	entities = map[string]Entity{}
	daos     = map[string]typeparser.Type{}
)

type Entity struct {
	Name          string
	TableName     string
	PrimaryKeys   map[string]string //field => column
	Attributes    map[string]string //field => column
	Options       map[string]typeparser.List
	AutoIncrement bool
	Type          typeparser.Type
}

func do(filename string, out io.Writer) error {
	tt, err := typeparser.Parse(filename)
	if err != nil {
		return err
	}
	if len(tt) == 0 {
		return fmt.Errorf("no types detected")
	}

	for _, t := range tt {
		if t.IsConcrete() {
			if len(t.Docs().Filter(func(s string) bool { return strings.HasPrefix(s, "//libdao") })) == 1 {
				e, err := processEntities(filename, t)
				if err != nil {
					return err
				}
				entities[t.Name()] = e
			}
		} else if t.IsInterface() && strings.HasSuffix(t.Name(), "DAO") {
			daos[t.Name()] = t
		}
	}

	for _, dao := range daos {
		if err := processDAOs(filename, dao, out); err != nil {
			return err
		}
	}

	return nil
}

func processEntities(filename string, t typeparser.Type) (Entity, error) {
	log.Printf("processing %s from %s", t.Name(), filename)
	e := Entity{
		Name:        t.Name(),
		Type:        t,
		PrimaryKeys: map[string]string{},
		Attributes:  map[string]string{},
		Options:     map[string]typeparser.List{},
	}

	annotation := t.Docs().Filter(func(s string) bool { return strings.HasPrefix(s, "//libdao") })[0]
	i := strings.LastIndex(annotation, "table=")
	e.TableName = annotation[i+6:]
	for _, f := range t.Fields() {
		if f.TagValue("db").Has("primarykey") {
			e.PrimaryKeys[f.Name()] = f.TagValue("db")[0]
			e.AutoIncrement = f.TagValue("db").Has("autoincrement")
		} else {
			e.Attributes[f.Name()] = f.TagValue("db")[0]
		}
		e.Options[f.Name()] = f.TagValue("db").Filter(
			func(s string) bool {
				return s != "primarykey" && s != "autoincrement"
			},
		)[1:]
	}
	return e, nil
}

func processDAOs(filename string, t typeparser.Type, out io.Writer) error {
	log.Printf("processing %s from %s", t.Name(), filename)

	target := t.Name()[:len(t.Name())-3]
	targetType, ok := entities[target]
	if !ok {
		return fmt.Errorf("entities %s not found", target)
	}

	fname := strings.ToLower(t.Name())
	f := jen.NewFile(fname)

	daoName := strings.ToLower(t.Name()[:1]) + t.Name()[1:]
	daoStruct := jen.Type().Id(daoName).Struct(
		jen.Id("db").Qual("github.com/fsamin/libdao", "SqlExecutor"),
		jen.Id("dbCtx").Qual("github.com/fsamin/libdao", "SqlExecutorWithContext"),
	)
	f.Add(daoStruct)

	for _, m := range t.Methods() {
		f.Add(
			processDAOFunc(filename, t, m, daoName, targetType),
		)
	}

	fmt.Fprintf(out, "%#v", f)

	return nil
}

func processDAOFunc(filename string, t typeparser.Type, m typeparser.Method, daoName string, targetType Entity) *jen.Statement {
	log.Println(m.Name())
	switch m.Name() {
	case "Insert":
		fun := methodOnType(m.Name(), daoName, targetType.Name)
		fun.Block(
			insertBlock(daoName, targetType),
			jen.Return(jen.Nil()),
		)
		return fun
	case "Update", "Delete":
		fun := methodOnType(m.Name(), daoName, targetType.Name)
		fun.Block(
			jen.Return(jen.Nil()),
		)
		return fun
	case "FindByID":
	}
	return nil
}

func methodOnType(name, receiver string, target string) *jen.Statement {
	fun := jen.Func()
	fun.Add(pointerReceiver("receiver", receiver))

	params := fun.Add(jen.Id(name))
	params = params.Add(jen.Parens(pointer("target", target)))
	params.Add(err())
	return fun
}

func pointerReceiver(id, name string) *jen.Statement {
	return jen.Parens(
		pointer(id, name),
	)
}

func pointer(id, name string) *jen.Statement {
	return jen.Id(id).Op("*").Qual("", name)
}

func err() *jen.Statement {
	return jen.Qual("", "error")
}

func insertBlock(daoName string, target Entity) *jen.Statement {
	/**
			q := "INSERT INTO my_type (\"bar\",\"biz\",\"foo\") VALUE ($1,$2,$3) RETURNING id"
		    if err := receiver.db.QueryRow(q, target.Bar, target.Biz, target.Foo).Scan(&target.ID); err != nil {
		            return err
		    }
			return nil

	or

			q := "INSERT INTO my_type (\"id\",\"bar\",\"biz\",\"foo\") VALUE ($1,$2,$3,$4) "
	        if err := receiver.db.Exec(q, target.ID, target.Bar, target.Biz, target.Foo); err != nil {
	                return err
	        }
	        return nil
	*/
	qattr := []string{}
	cattr := []jen.Code{}
	qvars := []string{}
	ckeys := []jen.Code{}
	var i = 1
	var qreturning string
	if !target.AutoIncrement {
		for f, c := range target.PrimaryKeys {
			qattr = append(qattr, `"`+c+`"`)
			qvars = append(qvars, "$"+strconv.Itoa(i))
			cattr = append(cattr, jen.Id("target").Dot(f))
			i++
		}
	} else {
		pkeys := []string{}
		for f, c := range target.PrimaryKeys {
			pkeys = append(pkeys, c)
			ckeys = append(ckeys, jen.Op("&").Id("target").Dot(f))
		}
		qreturning = fmt.Sprintf("RETURNING %s", strings.Join(pkeys, ","))

	}
	for f, c := range target.Attributes {
		qattr = append(qattr, `"`+c+`"`)
		qvars = append(qvars, "$"+strconv.Itoa(i))
		cattr = append(cattr, jen.Id("target").Dot(f))
		i++
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) %s", target.TableName, strings.Join(qattr, ","), strings.Join(qvars, ","), qreturning)
	qdecl := jen.Id("q").Op(":=").Lit(q)

	cattr = append([]jen.Code{jen.Id("q")}, cattr...)
	if target.AutoIncrement {
		qdecl.Line().Add(
			jen.If(
				jen.Id("err").Op(":=").Id("receiver").
					Dot("db").Dot("QueryRow").Call(
					cattr...,
				).Dot("Scan").Call(
					ckeys...,
				).Op(";").Id("err").Op("!=").Nil(),
			).Block(
				jen.Return(jen.Id("err")),
			),
		)
	} else {
		qdecl.Line().Add(
			jen.If(
				jen.Id("err").Op(":=").Id("receiver").
					Dot("db").Dot("Exec").Call(
					cattr...,
				).Op(";").Id("err").Op("!=").Nil(),
			).Block(
				jen.Return(jen.Id("err")),
			),
		)
	}

	return qdecl
}
