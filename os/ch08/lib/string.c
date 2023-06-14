#include "string.h"
#include "debug.h"
#include "global.h"

/**
 * 以字节为单位就是以字符为单位，内存是一个字节对应一个内存地址编号的，
 * 比如一个char a 或者一个uint8_t a。a均代表八位比特的数值（或ASCII码）。 
 * *a代表这个ASCII码的内存地址。++(*a)代表a的下一个内存地址编号
 */
/* 将dst_起始的size个字节置为value */
void memset(void* dst_,uint8_t value,uint32_t size)
{
    ASSERT(dst_ != NULL);
    uint8_t* dst = (uint8_t*) dst_;
    while((size--) > 0)
    	*(dst++) = value;
    return;
}

/* 将src_起始的size个字节复制到dst_ */
void memcpy(void* dst_,const void* src_,uint32_t size)
{
    ASSERT(dst_ != NULL && src_ != NULL);
    uint8_t* dst = dst_;
    const uint8_t* src = src_;
    while((size--) > 0)
    	*(dst++) = *(src++);
    return;
}

/* 连续比较以地址a_和地址b_开头的size个字节,若相等则返回0,若a_大于b_返回+1,否则返回-1 */
int memcmp(const void* a_,const void* b_, uint32_t size)
{
    const char* a = a_;
    const char* b = b_;
    ASSERT(a != NULL || b != NULL);
    while((size--) > 0)
    {
    	if(*a != *b)	return (*a > *b) ? 1 : -1;
   	++a,++b;
    }
    return 0;
}

/* 将字符串从src_复制到dst_ */
char* strcpy(char* dsc_,const char* src_)
{
    ASSERT(dsc_ != NULL && src_ != NULL);
    char* dsc = dsc_;
    while((*(dsc_++) = *(src_++) ));
    return dsc;     
}

/* 返回字符串长度 */
uint32_t strlen(const char* str)
{
    ASSERT(str != NULL);
    const char* ptr = str;
    while(*(ptr++));
    return (ptr - str - 1);             //例如一个字 1 '\0' ptr会指向'\0'后面一位
}

/* 比较两个字符串,若a_中的字符大于b_中的字符返回1,相等时返回0,否则返回-1. */
int8_t strcmp(const char* a,const char* b)
{
    ASSERT(a != NULL && b != NULL);
    while(*a && *a == *b)
    {
    	a++,b++;
    }
	/* 如果*a小于*b就返回-1,否则就属于*a大于等于*b的情况。在后面的布尔表达式"*a > *b"中,
     * 若*a大于*b,表达式就等于1,否则就表达式不成立,也就是布尔值为0,恰恰表示*a等于*b */
    return (*a < *b) ? -1 : (*a > *b) ;
}

/* 从左到右查找字符串str中首次出现字符ch的地址(不是下标,是地址) */
char* strchr(const char* str,const char ch)
{
    ASSERT(str != NULL);
    while(*str)
    {
    	if(*str == ch)	return (char*)str; // 需要强制转化成和返回值类型一样,否则编译器会报const属性丢失,下同.
    	++str;
    } 
    return NULL;
}

/* 从后往前查找字符串str中首次出现字符ch的地址(不是下标,是地址) */
char* strrchr(const char* str,const uint8_t ch)
{
    ASSERT(str != NULL);
    char* last_chrptr = NULL;
	/* 从头到尾遍历一次,若存在ch字符,last_char总是该字符最后一次出现在串中的地址(不是下标,是地址)*/
    while(*str != 0)
    {
    	if(ch == *str)	last_chrptr = (char*)str;
    	str++;
    }
    return last_chrptr;   
}

/* 将字符串src_拼接到dst_后,将回拼接的串地址 */
char* strcat(char* dsc_,const char* src_)
{
    ASSERT(dsc_ != NULL && src_ != NULL);
    char* str = dsc_;
    while(*(str++));
    str--;
    while(*(str++) = *(src_++));
    return dsc_;
}

/* 在字符串str中查找指定字符ch出现的次数 */
uint32_t strchrs(const char* str,uint8_t ch)
{
    ASSERT(str != NULL);
    uint32_t ch_cnt = 0;
    while(*str)
    {
    	if(*str == ch) ++ch_cnt;
    	++str;
    }
    return ch_cnt;
}