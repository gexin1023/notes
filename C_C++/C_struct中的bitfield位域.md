## C struct中的位域 bitfield

结构体的成员可以限制其位域，每个成员可以使用用比字节还小的取值范围，下面的结构体s1中，四个成员每个成员都是2bit的值（0~3），整个结构体占据的空间依然是4个字节，但是第一个字节中表示了四个成员，后续三个字节没有用到。

```c
struct {
    unsigned char a : 2;
    unsigned char b : 2;
    unsigned char c : 2;
    unsigned char d : 2;
} s1;

s1.a = 1;
s1.b = 1;
s1.c = 1;
s1.d = 1;

| 低位  ------>>>>>>>  高位 | 
|  byte 0  ||  byte 1  ||  byte 2  ||  byte 3  |
| 10101010 || 00000000 || 00000000 || 00000000 |

```

位域限制对于一些非字节对齐的变量，比较有用。有些标志位使用几个个bit就可以表示，这时可以用位域限制。我们以蓝牙mesh中的Network-PDU为例说明。

![Network_PDU](pic\Network_PDU.png)

IVI只占了1 bit，NID占了7bit。我们可以直接用1个字节表示，然后通过移位运算来表示IVI及NID。也可以使用位域，这样表示更加直接，与正常的结构体成员一样。

```c
// 使用位域
struct {
    unsigned char IVI : 1;
    unsigned char NID : 7;
    unsigned char CTL : 1;
    unsigned char TTL : 7;
} net_pdu;

net_pdu.IVI = 1;
net_pdu.NID = 123;

// 使用掩码
struct {
    unsigned char IVI_NID ;
    unsigned char CTL_TTL;
} net_pdu;

net_pdu.IVI_NID |= (1<<7)；
net_pdu.IVI_NID = (123) |  ((1<<7)&(net_pdu.IVI_NID));
```