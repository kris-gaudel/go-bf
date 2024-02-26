package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/marstr/collection/v2"
)

const (
	MemSize = 30000
)

type Loop struct {
	Body *ir.Block
	End  *ir.Block
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	program, err := io.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	mod := ir.NewModule()

	getChar := mod.NewFunc("getchar", types.I8)
	putChar := mod.NewFunc("putchar", types.I8, ir.NewParam("c", types.I8))
	memset := mod.NewFunc("memset", types.Void, ir.NewParam("ptr", types.I8Ptr), ir.NewParam("val", types.I8), ir.NewParam("len", types.I64))

	entryPoint := mod.NewFunc("main", types.I32)
	builder := entryPoint.NewBlock("")
	st := collection.NewStack[Loop]()

	// Init bf array
	arrayType := types.NewArray(MemSize, types.I8)
	bfArray := builder.NewAlloca(arrayType)
	dataPtr := builder.NewAlloca(types.I64)
	builder.NewStore(constant.NewInt(types.I64, 0), dataPtr)

	builder.NewCall(memset,
		builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), constant.NewInt(types.I64, 0)),
		constant.NewInt(types.I8, 0),
		constant.NewInt(types.I64, MemSize),
	)

	for i := 0; i < len(program); i++ {
		switch program[i] {
		case '+':
			ptr := builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), builder.NewLoad(types.I64, dataPtr))
			added := builder.NewAdd(builder.NewLoad(types.I8, ptr), constant.NewInt(types.I8, 1))
			builder.NewStore(added, ptr)
		case '-':
			ptr := builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), builder.NewLoad(types.I64, dataPtr))
			added := builder.NewAdd(builder.NewLoad(types.I8, ptr), constant.NewInt(types.I8, -1))
			builder.NewStore(added, ptr)
		case '>':
			t1 := builder.NewAdd(builder.NewLoad(types.I64, dataPtr), constant.NewInt(types.I64, 1))
			builder.NewStore(t1, dataPtr)
		case '<':
			t1 := builder.NewAdd(builder.NewLoad(types.I64, dataPtr), constant.NewInt(types.I64, -1))
			builder.NewStore(t1, dataPtr)
		case '.':
			ptr := builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), builder.NewLoad(types.I64, dataPtr))
			builder.NewCall(putChar, builder.NewLoad(types.I8, ptr))
		case ',':
			char := builder.NewCall(getChar)
			ptr := builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), builder.NewLoad(types.I64, dataPtr))
			builder.NewStore(char, ptr)
		case '[':
			ptr := builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), builder.NewLoad(types.I64, dataPtr))
			ld := builder.NewLoad(types.I8, ptr)
			cmpResult := builder.NewICmp(enum.IPredNE, ld, constant.NewInt(types.I8, 0))
			wb := Loop{
				Body: entryPoint.NewBlock(""),
				End:  entryPoint.NewBlock(""),
			}
			st.Push(wb)
			builder.NewCondBr(cmpResult, wb.Body, wb.End)
			builder = wb.Body
		case ']':
			front, ok := st.Pop()
			if !ok {
				log.Panic("unmatched ]")
			}
			ptr := builder.NewGetElementPtr(arrayType, bfArray, constant.NewInt(types.I64, 0), builder.NewLoad(types.I64, dataPtr))
			ld := builder.NewLoad(types.I8, ptr)
			cmpResult := builder.NewICmp(enum.IPredNE, ld, constant.NewInt(types.I8, 0))
			builder.NewCondBr(cmpResult, front.Body, front.End)
			builder = front.End
		}
	}

	builder.NewRet(constant.NewInt(types.I32, 0))

	fmt.Println(mod.String())
}
