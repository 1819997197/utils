#!/bin/bash

echo "中断处理程序"

nasm -I include/ -o mbr.bin mbr.S
dd if=/data/os/ch07/mbr.bin of=/data/os/c.img bs=512 count=1 conv=notrunc

nasm -I include/ -o loader.bin loader.S
dd if=/data/os/ch07/loader.bin of=/data/os/c.img bs=512 count=3 seek=2 conv=notrunc

gcc -m32 -I lib/kernel -c -o build/timer.o device/timer.c
gcc -m32 -I lib/kernel/ -m32 -I lib/ -m32 -I kernel/ -c -fno-builtin -o build/main.o kernel/main.c
gcc -m32 -I lib/kernel/ -m32 -I lib/ -m32 -I kernel/ -c -fno-builtin -o build/interrupt.o kernel/interrupt.c
gcc -m32 -I lib/kernel/ -m32 -I lib/ -m32 -I kernel/ -c -fno-builtin -o build/init.o kernel/init.c
nasm -f elf -o build/print.o lib/kernel/print.S
nasm -f elf -o build/kernel.o kernel/kernel.S
ld -m elf_i386 -Ttext 0xc0001500 -e main -o build/kernel.bin build/main.o build/init.o build/interrupt.o  build/print.o build/kernel.o build/timer.o
dd if=/data/os/ch07/build/kernel.bin of=/data/os/c.img bs=512 count=200 seek=9 conv=notrunc
