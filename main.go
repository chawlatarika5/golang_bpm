package main 

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Block struct {
	Index int
	Timestamp string
	BPM int
	Hash string
	PrevHash string
}

var Blockchain []Block

func calculateHash(block Block) string {
	record:= string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h:= sha256.New() //:= this is short version of declaring a var only "inside" the function
	h.Write([]byte(record))
	hashed:= h.Sum(nil)
	return hex.EncodeToString(hashed) //convert SHA byte int string
}

func generateBlock(oldBlock Block, BPM int) (Block, error) {
	var newBlock Block
	t := time.Now() //current time

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}


func isBlockValid(newBlock Block, oldBlock Block) bool {
	if (oldBlock.Index + 1 != newBlock.Index) {
		return false
	}
	if (oldBlock.Hash != newBlock.PrevHash) {
		return false
	}
	if (calculateHash(newBlock) != newBlock.Hash) {
		return false
	}
	return true
}

//overwrite the chain when two or more chains are available
//compare the length of the slices of the chain
//Blockchain is our var of array of blocks line 26

func replaceChain(newBlocks []Block) {
	if(len(newBlocks) > len(Blockchain)) {
		Blockchain = newBlocks
	}
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR") //coming from the 
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr: 			":" + httpAddr,
		Handler:		mux,
		ReadTimeout:	10 * time.Second,
		WriteTimeout:	10 * time.Second,
		MaxHeaderBytes:	1 << 20,
	}
	if err:= s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POSt")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//returns the blockchain in the json format
	io.WriteString(w, string(bytes))
}

//technically the user should be only adding data, everything else should be handled on its own
type Message struct {
	BPM int
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request){
	var m Message
	//converting the requested body into the message struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	//defer postpone the return until other functions return
	//in order to get the old block: we can get the (Blockchian -1) index
	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return 
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain) //useful for debugging
	}
	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(run())
}










