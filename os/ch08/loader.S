%include "boot.inc"
SECTION loader vstart=LOADER_BASE_ADDR
LOADER_STACK_TOP equ LOADER_BASE_ADDR 		           ; 是个程序都需要有栈区 设置的0x600以下的区域到0x500区域都是可用空间
jmp loader_start                     		   	       ; 下面存放数据段 构建gdt 跳跃到下面的代码区 
    GDT_BASE          dd 0x00000000          		   ; 刚开始的段选择子0不能使用 故用两个双字 来填充
   		              dd 0x00000000 
    CODE_DESC         dd 0x0000FFFF         		   ; FFFF是与其他的几部分相连接 形成0XFFFFF段界限
    		          dd DESC_CODE_HIGH4
    DATA_STACK_DESC   dd 0x0000FFFF
  		              dd DESC_DATA_HIGH4		       
    VIDEO_DESC        dd 0x80000007         		   ; 0xB8000 到0xBFFFF为文字模式显示内存
                      dd DESC_VIDEO_HIGH4     	       ; 0x0007 (bFFFF-b8000)/4k = 0x7
    GDT_SIZE          equ $ - GDT_BASE                 ; 当前位置减去GDT_BASE的地址 等于GDT的大小
    GDT_LIMIT         equ GDT_SIZE - 1   	           ; SIZE - 1即为最大偏移量 
    times 59          dq 0                             ; 预留59个 define double四字型 8字描述符
    times 5           db 0
    total_mem_bytes   dd 0                             ; 程序内偏移量为0x200 那这就是0x800
    gdt_ptr           dw GDT_LIMIT			           ; gdt指针 2字gdt界限放在前面 4字gdt地址放在后面 lgdt 48位格式 低位16位界限 高位32位起始地址
    		          dd GDT_BASE		       
    ards_buf times 244 db 0                            ; buf  记录内存大小的缓冲区
    ards_nr            dw 0					           ; nr 记录20字节结构体个数  计算了一下 4+2+4+244+2=256 刚好256字节							  
    SELECTOR_CODE     equ (0X0001<<3) + TI_GDT + RPL0  ; 16位寄存器 4位TI RPL状态 GDT剩下的选择子
    SELECTOR_DATA	  equ (0X0002<<3) + TI_GDT + RPL0
    SELECTOR_VIDEO    equ (0X0003<<3) + TI_GDT + RPL0   
    
loader_start:
    mov sp,LOADER_BASE_ADDR                            ; 先初始化了栈指针
    xor ebx,ebx                                        ; 异或自己 即等于0
    mov ax,0                                       
    mov es,ax
    mov di,ards_buf                                    ; di指向缓冲区位置
.e820_mem_get_loop:
    mov eax,0x0000E820
    mov ecx,0x14
    mov edx,0x534d4150
    
    int 0x15                                           ; 调用了0x15中断
    jc  .e820_failed_so_try_e801                       ; CF位=1 carry flag = 1 中途失败了即跳转
    add di,cx							               ; 把di的数值增加20 为了下一次作准备
    inc word [ards_nr]
    cmp ebx,0
    jne .e820_mem_get_loop                             ; 直至读取完全结束 则进入下面的处理时间
    
    mov cx,[ards_nr]
    mov ebx,ards_buf
    xor edx,edx
.find_max_mem_area:
    mov eax,[ebx]
    add eax,[ebx+8]
    add ebx,20
    cmp edx,eax
    jge .next_ards
    mov edx,eax

.next_ards:
    loop .find_max_mem_area
    jmp .mem_get_ok
    
.e820_failed_so_try_e801:                              ; 地址段名字取的真的简单易懂 哈哈哈哈 
    mov ax,0xe801
    int 0x15
    jc .e801_failed_so_try_88
   
;1 先算出来低15MB的内存    
    mov cx,0x400
    mul cx                                             ; 低位放在ax 高位放在了dx
    shl edx,16                                         ; dx把低位的16位以上的书往上面抬 变成正常的数
    and eax,0x0000FFFF                                 ; 把除了16位以下的 16位以上的数清零 防止影响
    or edx,eax                                         ; 15MB以下的数 暂时放到了edx中
    add edx,0x100000                                   ; 加了1MB 内存空缺 
    mov esi,edx
    
;2 接着算16MB以上的内存 字节为单位
    xor eax,eax
    mov ax,bx
    mov ecx,0x10000                                    ; 0x10000为64KB  64*1024  
    mul ecx                                            ; 高32位为0 因为低32位即有4GB 故只用加eax
    mov edx,esi
    add edx,eax
    jmp .mem_get_ok
 
.e801_failed_so_try_88:
     mov ah,0x88
     int 0x15
     jc .error_hlt
     and eax,0x0000FFFF
     mov cx,0x400                                      ; 1024
     mul cx
     shl edx,16
     or edx,eax 
     add edx,0x100000

.error_hlt:
     jmp $
.mem_get_ok:
     mov [total_mem_bytes],edx
; --------------------------------- 设置进入保护模式 -----------------------------
; 1 打开A20 gate
; 2 加载gdt
; 3 将cr0 的 pe位置1   
    in al,0x92                                         ; 端口号0x92 中 第1位变成1即可
    or al,0000_0010b
    out 0x92,al
   
    
    lgdt [gdt_ptr]
    
    
    mov eax,cr0                                        ; cr0寄存器第0位设置位1
    or  eax,0x00000001              
    mov cr0,eax
      
;-------------------------------- 已经打开保护模式 ---------------------------------------
    jmp dword SELECTOR_CODE:p_mode_start               ; 刷新流水线
 
 [bits 32]
 p_mode_start: 
    mov ax,SELECTOR_DATA
    mov ds,ax
    mov es,ax
    mov ss,ax
    mov esp,LOADER_STACK_TOP
    
;------------------------------- 加载内核到缓冲区 -------------------------------------------------
    mov eax, KERNEL_BIN_SECTOR
    mov ebx, KERNEL_BIN_BASE_ADDR
    mov ecx,200
    call rd_disk_m_32

;------------------------------- 启动分页 ---------------------------------------------------
    call setup_page
    sgdt [gdt_ptr]                                     ; 将gdt寄存器中的值 还是放到gdt_ptr内存中 我们修改相对应的 段描述符
    mov ebx,[gdt_ptr+2]                                ; 32位内存先倒出来 为的就是先把显存区域描述法的值改了
                                                               
    or dword [ebx+0x18+4],0xc0000000                   ; 段基址的最高位在高4字节
    add dword [gdt_ptr+2],0xc0000000                   ; gdt起始地址增加
    
    add esp,0xc0000000                                 ; 栈指针也进入高1GB虚拟内存区
    
    mov eax,PAGE_DIR_TABLE_POS
    mov cr3,eax
    
    mov eax,cr0
    or eax,0x80000000
    mov cr0,eax
    
    lgdt [gdt_ptr]
    
    mov eax,SELECTOR_VIDEO
    mov gs,eax
    mov byte [gs:160],'V'
    
    jmp SELECTOR_CODE:enter_kernel
    
;------------------------------ 跳转到内核区    
enter_kernel:
    call kernel_init					               ; 根据我们的1M以下的内存分布区 综合考虑出的数据
    mov  esp,0xc009f000
    jmp  KERNEL_ENTER_ADDR

;------------------------------- 创建页表 ------------------------------------------------    
setup_page:
    mov ecx,0x1000                                     ; 循环4096次 将页目录项清空 内存清0
    mov esi,0                                                   
 .clear_page_dir_mem:                                  ; dir directory 把页目录项清空
    mov byte [PAGE_DIR_TABLE_POS+esi],0
    inc esi
    loop .clear_page_dir_mem
    
 .create_pde: 
    mov eax,PAGE_DIR_TABLE_POS				           ; 页目录项 起始位置
    add eax,0x1000                                     ; 页目录项刚好4k字节 add eax即得第一个页表项的地址
                                                       ; 接下来要做的是 把虚拟地址1M下和3G+1M 两部分的1M内存在页目录项中都映射到物理地址0-0XFFFFF
    or  eax, PG_P | PG_RW_W | PG_US_U                  ; 设置为PG_US_U 是因为init在用户进程 如果这里设置成US_S 这样子连进内核都进不去了
     
    mov [PAGE_DIR_TABLE_POS+0x0],eax                   ; 页目录项偏移0字节与偏移0xc00 对应0x 一条页目录项对应2^22位4MB 偏移由前10位*4字节得到
    mov [PAGE_DIR_TABLE_POS+0xc00],eax                        
    sub eax,0x1000      
    
    mov [PAGE_DIR_TABLE_POS+4092],eax                  ; 虚拟内存最后一个目录项 指向页目录表自身

;这里就创建了一页页表    
    mov eax,PAGE_DIR_TABLE_POS
    add eax,0x1000
    mov ecx,256
    mov esi,0
    mov ebx,PG_P | PG_RW_W | PG_US_U 
    
 .create_kernel_pte:           
    mov [eax+esi*4],ebx
    inc esi
    add ebx,0x1000
    loop .create_kernel_pte 
    
    
;这里对于我们这里填写的目录项所对应的页表 页表中我们还没填写的值
;为了实现 真正意义上的 内核空间被用户进程完全共享
;只是把页目录与页表的映射做出来了 
    mov eax,PAGE_DIR_TABLE_POS
    add eax,0x2000       					           ; eax此时处于第二个页表
    or  eax,PG_P | PG_RW_W | PG_US_U
    mov ebx,PAGE_DIR_TABLE_POS
    mov ecx,254						  
    mov esi,769
        
 .create_kernel_pde:
    mov [ebx+esi*4],eax
    inc esi
    add eax,0x1000
    loop .create_kernel_pde 
    
    ret            
    
;----------------------- 初始化内核 把缓冲区的内核代码放到0x1500区域 ------------------------------------------
kernel_init:
    xor eax,eax                                        ; 全部清零
    xor ebx,ebx
    xor ecx,ecx
    xor edx,edx
    
    mov ebx,[KERNEL_BIN_BASE_ADDR+28]
    add ebx,KERNEL_BIN_BASE_ADDR                       ; ebx当前位置为程序段表
    mov dx,[KERNEL_BIN_BASE_ADDR+42]		           ; 获取程序段表每个条目描述符字节大小
    mov cx,[KERNEL_BIN_BASE_ADDR+44]                   ;一共有几个段
    
     
 .get_each_segment:
    cmp dword [ebx+0],PT_NULL
    je .PTNULL                                         ; 空即跳转即可 不进行mem_cpy
     
    push dword [ebx+16]                                ; ebx+16在存储的数是filesz
                                                               
    mov eax,[ebx+4]                                            
    add eax,KERNEL_BIN_BASE_ADDR
    push eax                                           ; p_offset 在文件中的偏移位置    源位置         
    push dword [ebx+8]                                 ; 目标位置
     
    call mem_cpy
    add esp,12                                         ; 把三个参数把栈扔出去 等于恢复栈指针
    
 .PTNULL:
    add  ebx,edx                                       ; edx是一个描述符字节大小
    loop .get_each_segment                             ; 继续进行外层循环    
    ret
                                        
mem_cpy:
    cld                                                ; 向高地址自动加数字 cld std 向低地址自动移动
    push ebp                                           ; 保存ebp 因为访问的时候通过ebp 良好的编程习惯保存相关寄存器
    mov  ebp,esp 
    push ecx                                           ; 外层循环还要用 必须保存 外层eax存储着还有几个段
    
    mov edi,[ebp+8]                                    ; 目的指针 edi存储的是目的位置 4+4
    mov esi,[ebp+12]                                   ; 源指针 源位置 8+4
    mov ecx,[ebp+16]                                   ; 与 Movsb 互相搭配  12+4
    
    rep movsb                                          ; 一个一个字节复制
       
    pop ecx 
    pop ebp
    ret
    
;------------------------ rd_disk_m_32  ----------------------
rd_disk_m_32:
;1 写入待操作磁盘数
;2 写入LBA 低24位寄存器 确认扇区
;3 device 寄存器 第4位主次盘 第6位LBA模式 改为1
;4 command 写指令
;5 读取status状态寄存器 判断是否完成工作
;6 完成工作 取出数据
 
;;;;;;;;;;;;;;;;;;;;;
;1 写入待操作磁盘数
;;;;;;;;;;;;;;;;;;;;;
    mov esi,eax
    mov di,cx
    mov dx,0x1F2
    mov al,cl
    out dx,al
    mov eax,esi
    
;;;;;;;;;;;;;;;;;;;;;
;2 写入LBA 24位寄存器 确认扇区
;;;;;;;;;;;;;;;;;;;;;
    mov cl,0x8    
    mov dx,0x1F3
    out dx,al 
    mov dx,0x1F4
    shr eax,cl
    out dx,al
    mov dx,0x1F5
    shr eax,cl
    out dx,al
    
;;;;;;;;;;;;;;;;;;;;;
;3 device 寄存器 第4位主次盘 第6位LBA模式 改为1
;;;;;;;;;;;;;;;;;;;;;
    shr eax,cl
    and al,0x0f 
    or al,0xe0
    mov dx,0x1F6
    out dx,al

;;;;;;;;;;;;;;;;;;;;;
;4 向Command写操作 Status和Command一个寄存器
;;;;;;;;;;;;;;;;;;;;;
    mov dx,0x1F7
    mov ax,0x20
    out dx,al
    
;;;;;;;;;;;;;;;;;;;;;
;5 向Status查看是否准备好惹 
;;;;;;;;;;;;;;;;;;;;;
  .not_ready:     
    nop
    in al,dx
    and al,0x88
    cmp al,0x08
    jne .not_ready
    
;;;;;;;;;;;;;;;;;;;;;
;6 读取数据
;;;;;;;;;;;;;;;;;;;;;
    mov ax,di
    mov dx,256
    mul dx
    mov cx,ax
    mov dx,0x1F0
 .go_read_loop:
    in ax,dx
    mov [ebx],ax
    add ebx,2
    loop .go_read_loop
    ret

