package gorocksdb

// #include <stdlib.h>
// #include "rocksdb/c.h"
import "C"
import (
	"errors"
	"unsafe"
)

// OptimisticTransactionDB is a reusable handle to a RocksDB optimistic transactional database on disk, created by OpenOptimisticTransactionDb.
type OptimisticTransactionDB struct {
	c    *C.rocksdb_optimistictransactiondb_t
	name string
	opts *Options
}

// OpenOptimisticTransactionDb opens a database with the specified options.
func OpenOptimisticTransactionDb(opts *Options, name string) (*OptimisticTransactionDB, error) {
	var (
		cErr  *C.char
		cName = C.CString(name)
	)
	defer C.free(unsafe.Pointer(cName))
	db := C.rocksdb_optimistictransactiondb_open(
		opts.c, cName, &cErr)
	if cErr != nil {
		defer C.rocksdb_free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return &OptimisticTransactionDB{
		name: name,
		c:    db,
		opts: opts,
	}, nil
}

// OpenOptimisticTransactionDbColumnFamilies opens a database with the specified options.
func OpenOptimisticTransactionDbColumnFamilies(
	opts *Options,
	name string,
	cfNames []string,
	cfOpts []*Options,
) (*OptimisticTransactionDB, []*ColumnFamilyHandle, error) {

	numColumnFamilies := len(cfNames)
	if numColumnFamilies != len(cfOpts) {
		return nil, nil, errors.New("must provide the same number of column family names and options")
	}

	var (
		cErr  *C.char
		cName = C.CString(name)
	)
	defer C.free(unsafe.Pointer(cName))

	cNames := make([]*C.char, numColumnFamilies)
	for i, s := range cfNames {
		cNames[i] = C.CString(s)
	}
	defer func() {
		for _, s := range cNames {
			C.free(unsafe.Pointer(s))
		}
	}()

	cOpts := make([]*C.rocksdb_options_t, numColumnFamilies)
	for i, o := range cfOpts {
		cOpts[i] = o.c
	}

	cHandles := make([]*C.rocksdb_column_family_handle_t, numColumnFamilies)

	db := C.rocksdb_optimistictransactiondb_open_column_families(
		opts.c,
		cName,
		C.int(numColumnFamilies),
		&cNames[0],
		&cOpts[0],
		&cHandles[0],
		&cErr,
	)
	if cErr != nil {
		defer C.rocksdb_free(unsafe.Pointer(cErr))
		return nil, nil, errors.New(C.GoString(cErr))
	}

	cfHandles := make([]*ColumnFamilyHandle, numColumnFamilies)
	for i, c := range cHandles {
		cfHandles[i] = NewNativeColumnFamilyHandle(c)
	}

	return &OptimisticTransactionDB{
		name: name,
		c:    db,
		opts: opts,
	}, cfHandles, nil

}

// GetBaseDb returns the handle to the underlying DB instance.
func (db *OptimisticTransactionDB) GetBaseDb() *DB {
	baseDb := C.rocksdb_optimistictransactiondb_get_base_db(db.c)
	return &DB{
		name: db.name,
		c:    baseDb,
		opts: db.opts,
	}
}

// TransactionBegin begins a new transaction
// with the WriteOptions and TransactionOptions given.
func (db *OptimisticTransactionDB) TransactionBegin(
	opts *WriteOptions,
	transactionOpts *OptimisticTransactionOptions,
	oldTransaction *Transaction,
) *Transaction {
	if oldTransaction != nil {
		return NewNativeTransaction(C.rocksdb_optimistictransaction_begin(
			db.c,
			opts.c,
			transactionOpts.c,
			oldTransaction.c,
		))
	}

	return NewNativeTransaction(C.rocksdb_optimistictransaction_begin(
		db.c, opts.c, transactionOpts.c, nil))
}

// Close closes the database.
func (transactionDB *OptimisticTransactionDB) Close() {
	C.rocksdb_optimistictransactiondb_close(transactionDB.c)
	transactionDB.c = nil
}
