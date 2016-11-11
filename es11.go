package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Fields struct {
	FN string
	FT string
	FS int
}

type esrec struct {
	RECNUMB    string
	STNAME     string
	LINETYPE   string
	LINECODE   string
	AREAOFFSET string
	LINECODETO string
	OLDSTATUS  string
	NEWSTATUS  string
	TALKFLAGS  string
	CAUSE      string
	ISUPCAT    string
	ENDDATE    string
	ENDTIME    string
	DURATION   string
	SUBSTO     string
	SUBSFROM   string
	REDIRSUBS  string
	CONNSUBS   string
	TALKCOMM   string
}

var wd, sp string

func main() {

	switch len(os.Args) {
	case 2:
		wd = os.Args[1]
	case 3:
		wd = os.Args[1]
		sp = os.Args[2]
	default:
		wd = "."
		sp = ""
	}

	//wd = "/home/vaxx/.go/src/github.com/vaxx99/es11/tmp"
	os.Chdir(wd)
	if info, err := os.Stat(wd); err == nil && info.IsDir() {
		f, _ := ioutil.ReadDir(wd)
		for _, fn := range f {
			if !fn.IsDir() && ises(fn.Name()) {
				_, rn, rec := es11(fn.Name())
				f, err := os.OpenFile(fn.Name()+".ama", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				defer f.Close()
				if err != nil {
					panic(err)
				}
				for _, j := range rec {
					ama := Erec2Str(j, sp)
					_, err = f.WriteString(ama + "\n")
				}
				fmt.Printf("%13s %10d %8d\n", fn.Name(), fn.Size(), rn)
			}
		}
	}
}

func Erec2Str(sr esrec, sp string) string {
	var ama string
	ama = sr.RECNUMB + sp + sr.STNAME + sp + sr.LINETYPE + sp + sr.LINECODE + sp + sr.AREAOFFSET + sp + sr.LINECODETO + sp + sr.OLDSTATUS + sp + sr.NEWSTATUS + sp + sr.TALKFLAGS + sp + sr.CAUSE + sp + sr.ISUPCAT + sp + sr.ENDDATE + sp + sr.ENDTIME + sp + sr.DURATION + sp + sr.SUBSTO + sp + sr.SUBSFROM + sp + sr.REDIRSUBS + sp + sr.CONNSUBS + sp + sr.TALKCOMM + sp
	return ama
}

func ises(fn string) bool {
	f, _ := os.Open(fn)
	v, _ := Read(f, 2)
	a := int(v[0])
	b := 1900 + int(v[1])
	defer f.Close()
	if a == 3 && b == time.Now().Add(-24*time.Hour).Year() {
		return true
	}
	return false
}

func es11(fn string) (string, uint32, []esrec) {
	var Esrec []esrec
	f, e := Open(fn)
	defer f.Close()
	dt, rn, hb, rb, fd, e := head(f)

	if e != nil {
		panic(e)
	}
	f, e = Open(fn)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	v, _ := Read(f, int(hb))
	var rec esrec

	recVal := reflect.ValueOf(&rec).Elem()
	for i := 0; i < int(rn); i++ {
		v, e = Read(f, int(rb))
		if e != nil {
			panic(e)
		}
		sb := 1
		eb := 1
		for a, b := range fd {
			eb = sb + b.FS
			s := strings.Replace(string(v[sb:eb]), " ", "", -1)
			recVal.Field(a).SetString(s)
			sb = eb
		}
		Esrec = append(Esrec, rec)
	}
	return dt, rn, Esrec
}

func dates(rec *esrec) (string, string, string) {
	dt := rec.ENDDATE
	tm := rec.ENDTIME
	de := dt + " " + tm
	dr := s2i(rec.DURATION)
	te, _ := time.Parse("20060102 15:04:05", de)
	ts := te.Add(time.Second * time.Duration(dr) * -1)
	de = te.Format("20060102150405")
	ds := ts.Format("20060102150405")
	return ds, de, strconv.Itoa(dr)
}

func s2i(s string) int {
	s = strings.Replace(s, " ", "", -1)
	a, _ := strconv.Atoi(s)
	return a
}

func head(f *os.File) (string, uint32, uint16, uint16, []Fields, error) {
	v, e := Read(f, 1)
	//Date modified 3
	v, e = Read(f, 3)
	//YY
	yy := strconv.Itoa(1900 + int(v[0]))
	mm := dd(int(v[1]))
	dd := dd(int(v[2]))
	dt := dd + "." + mm + "." + yy
	//Number of records 4-7 (4)
	v, e = Read(f, 4)
	rn := binary.LittleEndian.Uint32(v)
	//Number of bytes in header 8-9 (2)
	v, e = Read(f, 2)
	hb := binary.LittleEndian.Uint16(v)
	//Number of bytes in the record 10-11 (2)
	v, e = Read(f, 2)
	rb := binary.LittleEndian.Uint16(v)
	//12-14 	3 bytes 	Reserved bytes.
	v, e = Read(f, 3)
	//15-27 	13 bytes 	Reserved for dBASE III PLUS on a LAN.
	v, e = Read(f, 13)
	//28-31 	4 bytes 	Reserved bytes.
	v, e = Read(f, 4)
	//32-n 	32 bytes 	Field descriptor array (the structure of this array is each shown below)
	var br int
	var fld []Fields

	for br != 13 {
		v, e = Read(f, 32)
		br = int(v[0])
		if br != 13 {
			fld = append(fld, Fields{string(v[0:11]), string(v[11:12]), int(v[16])})
		}
	}
	f.Seek(0, 0)
	return dt, rn, hb, rb, fld, e
}

//Open file
func Open(fn string) (*os.File, error) {
	file, e := os.Open(fn)
	if e != nil {
		log.Println("File open error:", e)
	}
	return file, e
}

//Read file
func Read(file *os.File, bt int) ([]byte, error) {
	data := make([]byte, bt)
	_, e := file.Read(data)
	if e != nil {
		log.Println("File open error:", e)
	}
	return data, e
}

func dd(d int) string {
	if d < 10 {
		return "0" + strconv.Itoa(d)
	}
	return strconv.Itoa(d)
}
