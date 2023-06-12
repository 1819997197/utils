#include "print.h"
#include "init.h"
#include "debug.h"
#include "string.h"

int main(void) {
   put_str("I am kernel\n");
   init_all();
   ASSERT(strcmp("bbb","bbb"));
   while(1);
}
