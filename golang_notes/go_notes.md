## go miscs

+ go代码是用包来组织的，每个包有一个或多个go文件组成，这些go文件文件放在一个文件夹中
+ 每个源文件开始都用一个package声明，指明本源文件属于哪个包
+ pakage声明后紧跟这导入其他包
+ 导入包之后，是构成源文件的变量、函数、类型生命等
+ go语言不需要在语句后家分号
+ import时，左括号‘(’要跟import在一行
+ 函数的的左花括号'{' 必须跟func关键词在一行

下面这段代码是一个完整的GO程序
```go
package main

import(
	"fmt"
)

func main(){
    fmt.Println("hello world")
}
```
### 命令行参数

os包中提供了一些函数与变量与操作系统打交道。os.Args变量中存储命令行参数。

os.Args是一个字符串slice变量（字符串数组），可以直接print该值

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Args)
	fmt.Println(os.Args[1])
}
```

### 找出重复行

在输入中找不重复的行

```go
// 打印输入中输入次数大于1的行
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	count := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)

	for input.Scan() {
		if(input.Text() == ""){
			break
		}
		count[input.Text()]++
	}

	for line, num := range count {
		if num > 1 {
			fmt.Printf("%d\t%s\n", num, line)
		}
	}
}
```

map存储一个键值对组合，提供常量时间的操作来存储、获取、测试集合中的元素。键是可以进行相等比较的任意类型，字符串是最常见的键类型。值可以是任意类型。

map的创建使用make语句

```go
	count := make(map[string]int)
```

bufio包可以有效的处理输入输出，bufio.Scanner可以读取输入，以行或者单词为间隔。

下面这行代码表示从标准输入进读取，每次调用input.Scan()读取下一行，并且将结尾的换行去掉；调用input.Text()获取读到的内容。
```go
	input := bufio.NewScanner(os.Stdin)
	for input.Scan(){
        fmt.Println(input.Text())
	}
```

除了从标准输入中处理，更为广泛的是从文件中的输入内容中处理，下面代码实现了从文件中读取

```go
// 打印输入中输入次数大于1的行
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	count := make(map[string]int)

	files := os.Args[1:]

	if len(files) == 0 {
		countLines(os.Stdin, count)
	} else {
		for _, arg := range files {
			f, err := os.Open(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "dup err: %v\n", err)
				continue
			}
			countLines(f, count)
			f.Close()
		}
	}

	for line, num := range count {
		if num > 1 {
			fmt.Printf("%d\t%s\n", num, line)
		}
	}
}

func countLines(f *os.File, count map[string]int) {
	input := bufio.NewScanner(f)

	for input.Scan() {
		if input.Text() == "" {
			break
		}
		count[input.Text()]++
	}
}

```

首先在命令行参数中读取文件名，然后用os.Open()打开，此处返回的值是（f, err），f 的类型是\*os.File的指针类型。这里跟c语言中的标准流一样。

```go
	// files是输入文件名数组
	_, files := os.Args[1:]
	
	// 对每一个输入的文件进行读取, arg是每个文件的名字
	// f是文件指针
	for _,arg := files {
        f, err := os.Open(arg)
        countLines(f, count)
        f.Close()
	}
	
```

还需要注意的是，map是对make创建的数据结构的引用。** 当一个map被传递给一个函数时，函数接收到这个引用的副本，所有调用函数对map改变时，调用者使用的map也会产生改变。**这类似与C语言中的指针调用。

上面读取文件的方式是“流”的模式读取，用法与c语言中的文件流类似。原理上来说，可以用这种方式处理大量数据。

一个可选方式是一次读取一整块数据到大块内存，一次性的分割所有行，后面使用这种方式处理该问题。

```go
// 打印输入中输入次数大于1的行
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	count := make(map[string]int)

	files := os.Args[1:]

	if len(files) == 0 {
		countLines(os.Stdin, count)
	} else {
		for _, arg := range files {
			data, err := ioutil.ReadFile(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "dup err: %v\n", err)
				continue
			}
			for _, line := range strings.Split(string(data), "\n") {
				count[line]++
			}
		}
	}

	for line, n := range count {
		if n > 1 {
			fmt.Printf("%d\t%s\n", n, line)
		}
	}

}

func countLines(f *os.File, count map[string]int) {
	input := bufio.NewScanner(f)

	for input.Scan() {
		if input.Text() == "" {
			break
		}
		count[input.Text()]++
	}
}

```

这里是一次将全部的文件数据读入，然后根据"\n"将数据分割为行。

```go
	//读取文件内容到data
	data,err := ioutil.ReadFile(file)
	
	//按行分割
	strings.Split(string(data), "\n")
    
```

### GIF动画

```go
package main

import (
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var palette = []color.Color{color.White, color.Black}

const (
	whiteIndex = 0
	blackIndex = 1
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Args) > 1 && os.Args[1] == "web" {
		handler := func(w http.ResponseWriter, r *http.Request) {
			lisajous(w)
		}
		http.HandleFunc("/", handler)
		log.Fatal(http.ListenAndServe("localhost:8000", nil))
		return
	}
	lisajous(os.Stdout)
}

func lisajous(out io.Writer) {
	const (
		cycles  = 5
		res     = 0.001
		size    = 100
		nframes = 64
		delay   = 8
	)
	freq := rand.Float64() * 3.0
	anim := gif.GIF{LoopCount: nframes}
	phase := 0.0
	for i := 0; i < nframes; i++ {
		rect := image.Rect(0, 0, 2*size+1, 2*size+1)
		img := image.NewPaletted(rect, palette)
		for t := 0.0; t < cycles*2*math.Pi; t += res {
			x := math.Sin(t)
			y := math.Sin(t*freq + phase)
			img.SetColorIndex(size+int(x*size+0.5), size+int(y*size+0.5), blackIndex)
		}
		phase += 0.1
		anim.Delay = append(anim.Delay, delay)
		anim.Image = append(anim.Image, img)
	}
	gif.EncodeAll(out, &anim)
}

```

