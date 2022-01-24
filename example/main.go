package main

import (
	"sync/atomic"

	"github.com/corvinFn/mongo"
	"github.com/globalsign/mgo/bson"
)

type Record struct {
	TaxCode string `bson:"tax_code"`
	Date    string `json:"date"`
}

func main() {
	bson.SetJSONTagFallback(true)
	err := mongo.Init("prod")
	if err != nil {
		panic(err)
	}

	coll := mongo.Gdc.Open("gdc_invoice_statistic", "declare")
	defer coll.Close()

	coll.SetBatch(1024)
	iter := coll.Find(bson.M{}).Select(bson.M{"tax_code": 1, "date": 1}).Iter()
	defer iter.Close()

	var (
		count  int32
		record Record
	)
	for iter.Next(&record) {
		println(&record, record.TaxCode, record.Date)

		// limit output
		curCount := atomic.AddInt32(&count, 1)
		if curCount >= 200 {
			break
		}
	}
}
