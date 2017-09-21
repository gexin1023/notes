## c中的static变量

- static变量分配在内存中的数据段，函数内部声明的static变量在函数调用结束时，依然保持在内存中，

``` 
#include<stdio.h>

int fun()
{
  static int count = 0;
  count++;
  return count;
}
  
int main()
{
  printf("%d ", fun());
  printf("%d ", fun());
  return 0;
}

/*******函数输出如下**********/

Process started >>>
1, 2 
<<< Process finished. (Exit code 0)

```

- static变量如果没有初始化的话，会被隐式初始化为0

```
#include <stdio.h>
int main()
{
    static int x;
    printf("x = %d \n", x);
}


/**函数输出**/

x = 0

```

- 静态变量只能被const类型的变量初始化，例如下面的函数会出错

```
#include<stdio.h>
int initializer(void)
{
    return 50;
}
  
int main()
{
    static int i = initializer();
    printf(" value of i = %d", i);
    getchar();
    return 0;
}

/********编译失败******/

Process started >>>
D:\static.c: In function 'main':
D:\static.c:9:20: error: initializer element is not constant
     static int i = initializer();
                    ^
<<< Process finished. (Exit code 1)

```