## C_struct中的长度可变数组(Flexible array member)

Flexible array member  is a feature introduced in the C99 standard of the C programming language (in particular, in section §6.7.2.1, item 16, page 103). It is a member of a struct, which is an array without a given dimension, and it must be the last member of such a struct, as in the following example:

```c
struct vectord {
    uint8_t  len;
    double   arr[]; // the flexible array member must be last
};
```

+ `arr[]`不占用结构体的存储空间，sizeof(strcut  vectord)的值为1
+ 变长数组必须是结构体的最后一个成员
+ 结构体变量相邻的连续存储空间是`arr[]`数组的内容
+ gcc中使用0长度的数组`arr[0]`来表示变长数组。

```c
struct vectord *allocate_vectord (size_t len) {
   struct vectord *vec = malloc(offsetof(struct vectord, arr) + len * sizeof(vec->arr[0]));

   if (!vec) {
       perror("malloc vectord failed");
       exit(EXIT_FAILURE);
   }

   vec->len = len;

   for (size_t i = 0; i < len; i++)
       vec->arr[i] = 0;

   return vec;
}
```

