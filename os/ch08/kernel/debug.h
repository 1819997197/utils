#ifndef __KERNEL_DEBUG_H
#define __KERNEL_DEBUG_H

void panic_spin(char* filename, int line, const char* func, const char* condition);

// __VA_ARGS__ 是预处理器所支持的专用标识符
#define PANIC(...) panic_spin(__FILE__, __LINE__, __func__, __VA_ARGS__)

/**
 * ASSERT 是在调试过程中用的，经过预处理器展开后，
 * 调用宏的地方越多，程序的体积越大，所以执行得越慢。
 * 因此不需要调试时，应该取消 ASSERT 
 */
#ifdef NDEBUG
    #define ASSERT(CONDITION) ((void) 0) // 取消 ASSERT
#else
    #define ASSERT(CONDITION) \
        if(CONDITION) {} else { \
            /* 符号“#”让编译器将宏的参数转化为字符串字面量, 例如：a==b 变成 "a==b" */ \
            PANIC(#CONDITION); \
        }
#endif

#endif