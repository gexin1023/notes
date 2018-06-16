## 程序结构

### 2.1 命名

Golang中的命名遵循这样一个简单原则，名字的开头必须是字母或者下划线，后面跟字母、数字或者下划线（这里与C语言中是一致的）。

在函数内部声明的实体，即局部变量，只在函数内部有效。在函数外定义的变量，在整个包内有效（注意是包，不是文件，多个文件可以属于同一个包）。

首字母的大小写决定了是否对其他包可见，首字母小写只对本包内有效，首字母大写对于其他包可见。比如，`fmt.Println()`函数名字的首字母大写，可以在其他包内引用该函数。

名字中出现多个单词时，习惯上使用大写单词首字母以便于阅读（一般不用下划线分割），比如`parseRequestLine()`。

### 2.2 声明

声明命名了一个程序实体，并制定了它的一些特性。有四种主要声明，分别是`var`、`const`、`type`、`func`。

go程序的结构遵循如下的结构:

```go
// package name
package main

// import other packages
import (
	"fmt"
)

type student struct {
	name string
	age  int
}

var (
	x int     = 0
	y float64 = 1.1
	z bool    = false
)

const (
	dayOfWeek = 7
)

func main() {
	gexin := student{name: "gexin", age: 27}

	fmt.Println(gexin, x, y, z, dayOfWeek)
}

```

### 2.3 变量

一个变量声明创建了一个特定类型的变量，并给变量一个名字，给变量赋初值。

```
var  name type = expression

var  x int  = 10
var  x      = 10
var  x int
```

声明变量时，`type`与`expression`至少存在一个。如果不存在type，则根据expression的类型来确定变量类型；如果不存在expression，则将其初始化为0。

    对于数值型变量其默认初始化为0，布尔变量默认初始化为false，字符串默认初始化为空字符串""，引用类型（slice, pointer, map, channel, function）默认初始化为nil，组合类型（array、struct）所有的成员都默认初始化为0.

0初始化机制使得任何变量始终都有一个正确的值（不像C语言中，对于未初始化的某些变量可能会造成问题，尤其是内存相关的)。

显式初始化的值可以是字符串值，亦可以是任意表达式。包级别的变量（在函数体外部声明的变量）实在main()函数开始之前初始化的，局部变量是在函数运行时在其声明的地方初始化的。

多个变量可以同时声明并用函数初始化，如下所示：

```go 
var f, err = os.Open(arg)
```

#### 2.3.1 短变量声明

在函数体内部，可以使用另一种声明格式，称之为"短变量声明"，采用如下的格式：

```go 
x := 10
y := 3.14
f, err := os.Open(arg)
```

多数情况下局部变量使用短变量声明，变量声明用在那些需要显示表明变量类型的地方，或者是变量初值不重要，只是需要一个类型的变量的地方。

短变量声明也支持多变量同时声明，如下所示：

```
i, j := 0, 1
```

短变量声明也可在函数调用时声明一个变量作为函数返回值，需要始终注意的是`:=`是声明，其左侧的多个变量**至少有一个**是新声明的变量，对于已存在的变量其作用相当于赋值。

```go 
f, err := os.Open(fileName)
```

#### 2.3.2 指针

变量是一段保存了一个值得存储空间，变量是在声明过程中创建的，并且给了一个名字用来访问变量。也有一些变量是通过表达式来访问的，比如`x[i]`、`x.f`，这类表达式读取变量的值，当出现在赋值号左边时，就是给变量赋值。

一个指针的值是一个变量的地址，指针是一个值在内存中存储的位置。不是每个值都有地址，但是每个变量都有地址，变量也可以成为可以被寻址的值。通过指针，我们可以访问或者修改某个变量的值（直接通过变量地址，不用知道变量名字）。

    “变量”与“值”这俩概念有点拗口，听起来有点别扭。变量是一段存储空间，里面的实际内容就是值。变量也可以成为可以被寻址的值。变量的内容可以通过变量名字访问，亦可以直接通过指针访问，指针就是变量的存储地址。

对于一个已经声明的变量`x int`，`&x`表示地址可以赋值给指针，这个地址值得类型是`*int`。

```go 
x := 1
p := &x

fmt.Println(*p) // 1 

*p=2
fmt.Println(x) // 2 
```

指针的默认初始化0值时`nil`，指针时可以比较的，当两个指针指向同一个变量时，他们是相等的。

在golang中，函数返回局部变量的地址是安全的（这一点比C好），比如下面的代码的代码中，在函数f1返回后，变量v的地址仍然是有效的，其值为10；

```go 
func main(){
    p := f1()
    fmt.Println(*p) // 10
}

func f1() *int {
	v := 10
	return &v
}
```

在函数调用中，亦可用传地址方式修改参数，函数的参数设置为指针类型，就可以了。这点与C语言一致。


#### 2.3.3 new()函数

new(T)会创建一个变量，0初始化并返回其地址，但是没有给这个变量名字。

```go 
p := new(int)   // p是*int类型的指针，*p值初始化为0
fmt.Println(*p) // 0

*p = 2
fmt.Println(*p) // 2
```
需要注意的是，通过new来创建的变量与普通创建的变量并无什么不同，只是没有给变量起名字罢了，所以下面代码段中的两个函数本质是一样的。

```go 
func newInt() *int{
    return new(int)
}
// **************************
func newInt() *int{
    var dummy int
    return &dummy
}
```

#### 2.3.4 变量的生命周期

包级别变量存在于程序的整个执行期间，局部变量有着动态的声明周期。局部变量在声明时被创建，直到变量不被访问才会被销毁。函数参数及返回结果也是局部变量，当函数调用时被创建。

```go 
for t := 0.0; t < cycles*2*math.Pi; t += res {
    x := math.Sin(t)
    y := math.Sin(t*freq + phase)
    img.SetColorIndex(size+int(x*size+0.5), size+int(y*size+0.5), blackIndex)
}
```
在上面的代码段中，t是在循环开始时被创建，x,y在每次循环中创建。

垃圾收集机制是如何判断何时收回变量呢？这是个比较复杂的问题，这里简单介绍一下其原理。每次创建一个变量时，该变量都作为一个根路径来被他的指针或者其他引用跟踪。如果这样的路径都不存在了，那么这个变量就不可访问了，这时就可以回收了。

变量的生命周期取决于它是否可以被访问，一个局部变量可以存在于他的代码段之外，所以函数返回局部变量的地址也是安全的。


### 2.4 赋值

变量的值是通过赋值语句来更新的，如下所示：

```go 
x = 1
*p = true
person.name = "gexin"
count[x] = count[x] * scale
```

跟C语言中一样，以下的形式也是正确的

```go
count[x] *= 2

v := 1
v++
v--
```

#### 2.4.1 Tuple（元组）赋值

元组赋值是说多个变量同时赋值

```go
// 交换变量的值
x, y = y, x
a[i], a[j] = a[j], a[i]
```

求两个整数的最大公约数
```go
func gdc(x, y int) int {
	for y != 0{
		x, y = y, x%y
	}
	return x
}
```

求第n个菲波那切数列

```go
func fib(n int) int{
	x, y := 0, 1
	for i:=0; i<n; i++{
		x, y = y, x+y
	}
	return x
}
```

一些函数需要返回额外的错误码以表明程序执行的状态，比如之前用的到`os.Open()`，这时就需要元组赋值了，如下所示：

```go
f, err = os.Open("file.txt")
```

有三个操作符有时也表现出相同的方式，如下所示：

```go
v, ok = m[key] 		// map lookup
v, ok = x.(T)		// type assertion
v, ok = <-ch 		// chanel receive
```
就像变量声明一样，我们也可以用下划线来赋值不想用的值，如下所示：

```go
_, err = io.Copy(dst, src)
_, ok  = x.(T)
```

#### 2.4.2 可赋值性

除了显示的赋值，还有很多地方会有隐式赋值。程序调用时，会隐式的给参数变量赋值；程序返回时，隐式的给结果变量赋值；符合结构的数据使用字符常量，默认给每个成员赋值，如下所示：

```go 
medals := []string{"gold", "silver", "bronze"}

// 对每个元素隐式赋值，相当于如下三个赋值
medals[0] = "gold"
medals[1] = "silver"
medals[2] = "bronze"
```

一个赋值操作，不管是显式的还是隐式的，只要两侧的类型一致，该操作就是合法的。

两个值是否相等，`==`或者`!=`，与可赋值性相关。在比较操作中，第一个操作数必须可以被第二个的数据类型赋值，反之亦然。

### 2.5 类型声明

变量或者表达式的类型决定了值得表现形式， 比如值得size，如何表示，支持的运算，与之关联的操作方法等等。

```go
type name underlying_type
```

一个类型声明定义了一个名为"name"的类型，它与 "underlying_type"有着相同的类型。

```go
type Celsius  float64
type student struct {
	 underlying_type
	age 	int
}name string	
	
age 	int
对于每个类型T，都有一个转换操作，T(x)，该操作将x值转换为T类型。转换在以下几种情况下才是允许的：

+ x的类型和T都有着相同的"underlying_type"
+ 都是未命名的指针类型，而且指向相同的"underlying_type"数据
+ 虽然改变的类型，但是不影响值得表达

转换在数值类型间是可以转换的，字符串和一些slice类型间也是可以转换的。这些转换可能会影响值得表达，比如将一个float64类型装换为int。

### 2.6 包和文件

go中的pacakge就跟C语言中的库是一样的，包的源码分布在一个或者多个.go文件中，每个包给其中的声明都提供了一个独立的命名空间，比如`utf16.Decode()`与`image.Decode()`就是两个不同的函数。

包用简单的方式决定一个变量是否可以被包外访问，首字母大写的才可以在包外访问，首字母小写的只能在包内访问。


我们在这里实现温度转换的例子，该包使用两个文件来实现，一个包用来声明类型、常量等信息，另一个包用来实现方法。

```go 
// file tempconv.go
package tempconv

import (
	"fmt"
)

type Celsius float64
type Faherenheit float64

const (
	AbsoluteZeroC Celsius = -273.15
	FreezingC     Celsius = 0
	BoilingC      Celsius = 100
)

func (c Celsius) String() string     { return fmt.Sprintf("%g°C".c) }
func (f Faherenheit) String() string { return fmt.Sprintf("%g°F", f) }

// file conv.go
package tempconv

func CToF(c Celsius) Faherenheit { return Faherenheit(c*9/5 + 32) }
func FToC(f Faherenheit) Celsius { return Celsius((f - 32) * 5 / 9) }
```
