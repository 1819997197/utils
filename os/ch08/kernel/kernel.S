[bits 32]
%define ERROR_CODE nop	 ; 若CPU压入了错误码，则无需操作
%define ZERO push 0		 ; 若CPU没有压入错误码，则需要手动压入一个0，保证格式统一

extern put_str;
extern idt_table;

section .data
global intr_entry_table  ; 存储各个中断的中断例程入口地址
intr_entry_table:

%macro VECTOR 2          ; %1:中断向量号, %2:对压入 ERROR_CODE 的处理
section .text
intr%1entry:		     ; 中断处理程序的起始处

   %2				     ; 中断若有错误码会压在eip后面 
   push ds               ; 以下是保存上下文环境
   push es
   push fs
   push gs
   pushad			     ; PUSHAD指令压入32位寄存器,其入栈顺序是: EAX,ECX,EDX,EBX,ESP,EBP,ESI,EDI

   ; 如果是从片上进入的中断,除了往从片上发送EOI外,还要往主片上发送EOI 
   mov al,0x20           ; 中断结束命令EOI
   out 0xa0,al           ; 向从片发送
   out 0x20,al           ; 向主片发送

   push %1
   call [idt_table + %1*4]  ; 调用idt_table中的C版本中断处理函数
   jmp intr_exit
	
section .data
   dd    intr%1entry	    ; 存储各个中断入口程序的地址，形成intr_entry_table数组
%endmacro

section .text
global intr_exit
intr_exit:	                ; 以下是恢复上下文环境
   add esp, 4			    ; 跳过中断号
   popad
   pop gs
   pop fs
   pop es
   pop ds
   add esp, 4			    ; 跳过error_code
   iretd

VECTOR 0x0 ,ZERO
VECTOR 0X1 ,ZERO
VECTOR 0X2 ,ZERO
VECTOR 0x3 ,ZERO
VECTOR 0X4 ,ZERO
VECTOR 0X5 ,ZERO
VECTOR 0x6 ,ZERO
VECTOR 0X7 ,ZERO
VECTOR 0X8 ,ERROR_CODE
VECTOR 0x9 ,ZERO
VECTOR 0XA ,ERROR_CODE
VECTOR 0XB ,ERROR_CODE
VECTOR 0XC ,ERROR_CODE
VECTOR 0XD ,ERROR_CODE
VECTOR 0XE ,ERROR_CODE
VECTOR 0XF ,ZERO
VECTOR 0X10 ,ZERO
VECTOR 0X11 ,ERROR_CODE
VECTOR 0x12 ,ZERO
VECTOR 0X13 ,ZERO
VECTOR 0X14 ,ZERO
VECTOR 0x15 ,ZERO
VECTOR 0X16 ,ZERO
VECTOR 0X17 ,ZERO
VECTOR 0X18 ,ZERO
VECTOR 0X19 ,ZERO
VECTOR 0X1A ,ZERO
VECTOR 0X1B ,ZERO
VECTOR 0X1C ,ZERO
VECTOR 0X1D ,ZERO
VECTOR 0X1E ,ERROR_CODE
VECTOR 0X1F ,ZERO
VECTOR 0X20 ,ZERO

