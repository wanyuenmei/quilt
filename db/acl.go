package db

import (
	"errors"
	"log"
)

// ACL defines access control for Quilt-managed machines.
type ACL struct {
	ID int

	Admin []string
}

// InsertACL creates a new ACL row and inserts it into 'db'.
func (db Database) InsertACL() ACL {
	result := ACL{ID: db.nextID()}
	db.insert(result)
	return result
}

// SelectFromACL gets all acls in the database that satisfy 'check'.
func (db Database) SelectFromACL(check func(ACL) bool) []ACL {
	result := []ACL{}
	for _, row := range db.tables[ACLTable].rows {
		if check == nil || check(row.(ACL)) {
			result = append(result, row.(ACL))
		}
	}
	return result
}

// GetACL gets the ACL row from the database. There should only ever be a single
// ACL row.
func (db Database) GetACL() (ACL, error) {
	aclRows := db.SelectFromACL(nil)
	numACLs := len(aclRows)
	if numACLs == 1 {
		return aclRows[0], nil
	} else if numACLs > 1 {
		log.Panicf("Found %d ACL rows, there should be 1", numACLs)
	}
	return ACL{}, errors.New("no ACL rows found")
}

func (acl ACL) getID() int {
	return acl.ID
}

func (acl ACL) tt() TableType {
	return ACLTable
}

func (acl ACL) String() string {
	return defaultString(acl)
}

func (acl ACL) less(r row) bool {
	return acl.ID < r.(ACL).ID
}
