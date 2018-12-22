<!---
title:: Pintos实验阶段总结4
date:: 2015-03-24 21:08
categories:: 系统与网络
tags:: c, pintos, memory
-->

最近这段时间，参考别人的方案，自己尝试完成了Pintos的Project 3，这个Project其实如果对虚拟内存机制比较熟悉，有相关经验，也没有想象中那么难。Project3的解决方案思路其实并不是很难，但是设计的结构体比较多，所以实现玩之后的debug工作更加难一些。我先将大概讲述Project 3的解决思路，然后给出几个测试例程的debug过程中的一些重要信息来说明虚拟内存的一些关键点。
## 解决思路
在[Pintos虚拟内存机制初探](./pintos-vm.md)的博文中，我也大致比较明白说明了虚拟内存的关键点，虚拟内存是操作系统内存管理模块和CPU的MMU硬件协同实现的。Pintos原始的代码中已经实现了虚拟内存，但是这里面的虚拟内存不够完善，没有实现内存页的换入换出操作，没有实现MMAP系统调用。这样的虚拟内存也仅仅是做了一个虚拟地址到物理地址的映射，保护程序之间不互相干扰而已。在Pintos的原始实现中，是一次性将User Program的整个文件load到内存中，然后建立一个大小固定为1Page的stack，这不是一个好的做法。

在Project 3中，我们主要有3个任务：
+ 在User Program的load过程中，仅仅是在User Program对应的thread结构体的spt链表中添加记录而已，并不实际load文件块到内存中，stack的大小不是固定的，可以随着以后的需要增长；
+ 将一块磁盘块设备作为内存页交换的SWAP分区，实现内存页的换入换出；
+ 实现MMAP和MUNMAP系统调用。
```
struct frame_entry {
    void *frame;
    struct sup_page_entry *spte;
    struct thread *thread;
    struct list_elem elem;
};

struct sup_page_entry {
    uint8_t type;
    void *uva;
    bool writable;

    bool is_loaded;
    bool pinned;

    // For files
    struct file *file;
    size_t offset;
    size_t read_bytes;
    size_t zero_bytes;

    // For swap
    size_t swap_index;

    struct hash_elem elem;
};
```
这里最重要的就是上面的2个数据结构。其中`frame_entry`以链表结构组织在一起，系统中有一个全局变量`struct list frame_table`，记录系统中正在被使用的内存页的相关信息，这个结构体是为了实现页的换入换出的。`sup_page_entry`是以hash table组织在一起的，每个thread结构体中都一个`struct hash spt`哈希表，这个哈希表上记录的就是这个thread的地址空间信息，虚拟地址的大部分操作都和这个结构体有关。

MMAP系统调用实质上也就是在thread的spt上添加文件和内存之间的映射信息，在移除的时候，看映射的页是否dirty，如果dirty，需要回写到磁盘上，否则直接移除MMAP的内存映射信息。用于SWAP交换空间的磁盘的管理和内存的管理类似，也是通过bitmap来实现的，页的换入换出操作从程序的地址空间的角度来看，其实和MMAP有些类似，也是内存和磁盘的映射，但是SWAP相当于一个内存页的临时存放的地方，在系统内存负载较大时，会频繁有页的换入换出操作。
```
/* Page fault handler.  This is a skeleton that must be filled in
   to implement virtual memory.  Some solutions to project 2 may
   also require modifying this code.

   At entry, the address that faulted is in CR2 (Control Register
   2) and information about the fault, formatted as described in
   the PF_* macros in exception.h, is in F's error_code member.  The
   example code here shows how to parse that information.  You
   can find more information about both of these in the
   description of "Interrupt 14--Page Fault Exception (#PF)" in
   [IA32-v3a] section 5.15 "Exception and Interrupt Reference". */

#define USER_VADDR_BOTTOM ((void *)0x08048000)
#define STACK_HEURISTIC 32

static void
page_fault (struct intr_frame *f)
{
    bool not_present;  /* True: not-present page, false: writing r/o page. */
    bool write;        /* True: access was write, false: access was read. */
    bool user;         /* True: access by user, false: access by kernel. */
    void *fault_addr;  /* Fault address. */

    /* Obtain faulting address, the virtual address that was
       accessed to cause the fault.  It may point to code or to
       data.  It is not necessarily the address of the instruction
       that caused the fault (that's f->eip).
       See [IA32-v2a] "MOV--Move to/from Control Registers" and
       [IA32-v3a] 5.15 "Interrupt 14--Page Fault Exception
       (#PF)". */
    asm ("movl %%cr2, %0" : "=r" (fault_addr));

    /* Turn interrupts back on (they were only off so that we could
       be assured of reading CR2 before it changed). */
    intr_enable ();

    /* Count page faults. */
    page_fault_cnt++;

    /* Determine cause. */
    not_present = (f->error_code & PF_P) == 0;
    write = (f->error_code & PF_W) != 0;
    user = (f->error_code & PF_U) != 0;

#ifdef VM
    struct thread *t=thread_current();
    void *esp=(f->cs==SEL_KCSEG) ? t->esp : f->esp;
    bool load=false;
    if(not_present && fault_addr > USER_VADDR_BOTTOM && is_user_vaddr(fault_addr))
    {
        struct sup_page_entry *spte=get_spte(fault_addr);
        if(spte)
        {
            load=load_page(spte);
            spte->pinned=false;
        }
        else if(fault_addr >= esp-STACK_HEURISTIC)
        {
            load=grow_stack(fault_addr);
        }

        if(!load)
        {
#endif

            thread_current()->exit_status=-1;
            thread_exit();
            kill (f);
#ifdef VM
        }
    }
#endif
}
```
这种情况下，系统中会经常出现page fault这个exception，所以我们要去修改`page_fault`这个函数，在出现缺页异常时，能够正确处理。`page_fault()`异常处理例程需要判断是属于哪一种情况，进行相应的处理。
## pt-grow-stk-sc调试记录
```
void
test_main (void)
{
    int handle;
    int slen = strlen (sample);
    char buf2[65536];

    /* Write file via write(). */
    CHECK (create ("sample.txt", slen), "create \"sample.txt\"");
    CHECK ((handle = open ("sample.txt")) > 1, "open \"sample.txt\"");
    CHECK (write (handle, sample, slen) == slen, "write \"sample.txt\"");
    close (handle);

    /* Read back via read(). */
    CHECK ((handle = open ("sample.txt")) > 1, "2nd open \"sample.txt\"");
    CHECK (read (handle, buf2 + 32768, slen) == slen, "read \"sample.txt\"");

    CHECK (!memcmp (sample, buf2 + 32768, slen), "compare written data against read data");
    close (handle);
```
上面是一个测试例程，这个例程最主要的是为了测试thread的stack即使在系统调用中也能够正确扩展。下面我们来捕捉程序中的page_fault和write read系统调用的过程，以及一些记录信息。
```
load()例程结束后：
if_.eip=0x0804881c(program entry)
if_.esp=0xc0000000

测试例程总共发生了8次page fault，下面是这8次page fault中fault_addr write user信息
注：write代表是否可写，user代表是内核还是用户所有
1.0x804881c F T
这个page fault地址就是程序的入口地址，是程序的代码段
2.0x804c5e4 T T
f->cs SEL_UCSEG
f->esp 0xbfffffa0
这个write为True，我不是很确定是那个段，可能是bss段吧
3.0x804948f F T
程序的代码段
4.0x804a872 F T
f->esp 0xbffffe90
5.0x804bd60 F T
f->esp 0xbffeff80
6.0xbffeff80 T T
f->cs SEL_UCSEG
f->esp 0xbffeff80
thread_current()->esp 0xbfffff3c
程序中write read系统调用依次发生在这里
7.0xbfff7f90 F F
f->cs SEL_KCSEG
f->esp 0xc002c0cb(syscall_handler+771)
thread_current()->esp 0xbffeff6c
8.0xbfff8000 T F
f->cs SEL_KCSEG
f->esp 0xc0038b14(channel+52)
thread_current()->esp 0xbffeff6c

进入SYS_WRITE系统调用的是：
write (handle, sample, slen)
sample是一个全局变量，slen是长度
在系统调用的代码处：
fd:3(handle) buffer:0x804bd60(sample) size:794(slen)
buffer的地址和第5次page fault的值相同，这次page发生在write系统调用之前
如果我没想错的话，应该是int slen = strlen (sample);触发了这次page fault的发生
sample是全局静态变量，所以其存储在只读数据区中，所以第5次page fault应该load的是这个只读数据区进来
f->esp 0xbffeff6c

进入SYS_READ系统调用的是：
read (handle, buf2 + 32768, slen)
buf2是例程中的局部变量，存储在Process的用户stack中
fd:4(handle) u_buffer 0xbfff7f90(buf2+32768) size:794(slen)
由于buf2是局部变量，动态存储在stack中，具有生存期，所以这些局部变量在load是不会分配静态存储空间的
这些和栈有关的由esp和ebp来指示，通过push和pop来操作，发生page_fault是，会导致stack的扩展
在这里，stack的扩展还不是连续的，而且还不能动态回收这些栈使用的页，除非程序退出才由操作系统回收
f->esp 0xbffeff6c
f->cs SEL_UCSEG

第7次page fault就是read系统调用后触发的，也就是直接从系统调用中进入page_fault()例程
此时page_fault()传入进来的intr_frame *f并不是用户空间的了，而是内核空间的
也就是说f->esp不是代表用户stack的栈顶了，而是内核栈的栈顶，所以我们要区分进入page_fault()例程是用户空间，还是内核空间，分别处理
为了解决这个问题，我在struct thread结构体中添加了一个esp字段，用于记录用户空间栈顶的信息
在进入系统调用后就更新这个esp字段值，然后在page_fault()例程中根据是用户空间还是内核空间进入使用哪个esp值
struct thread *t=thread_current();
void *esp=(f->cs==SEL_KCSEG) ? t->esp : f->esp;
在这次page fault中：
t->esp 0xbffeff6c
f->esp 0xc002c0cb
f->cs SEL_KCSEG
明显可以看到f->esp是内核栈栈顶

这里其实也有一个问题，第7次page fault中write为False，user也为False，证明这个页不可写，还属于内核
其实这不科学，但是能行就好
```
从上面这个调试例子，可以看到至少程序的load和stack的扩展是可以正确工作的，struct thread中的spt哈希表记录地址空间也是可以正确工作，page_fault也是基本正确的。
## mmap-inherit调试记录
```
/***********mmap-inherit.c***********************/
void
test_main (void)
{
    char *actual = (char *) 0x54321000;
    int handle;
    pid_t child;

    /* Open file, map, verify data. */
    CHECK ((handle = open ("sample.txt")) > 1, "open \"sample.txt\"");
    CHECK (mmap (handle, actual) != MAP_FAILED, "mmap \"sample.txt\"");
    if (memcmp (actual, sample, strlen (sample)))
        fail ("read of mmap'd file reported bad data");

    /* Spawn child and wait. */
    CHECK ((child = exec ("child-inherit")) != -1, "exec \"child-inherit\"");
    quiet = true;
    CHECK (wait (child) == -1, "wait for child (should return -1)");
    quiet = false;

    /* Verify data again. */
    CHECK (!memcmp (actual, sample, strlen (sample)),
            "checking that mmap'd file still has same data");
}

/***********child-inherit.c***********************/
void
test_main (void)
{
    memset ((char *) 0x54321000, 0, 4096);
    fail ("child can modify parent's memory mappings");
}
```
mmap-inherit测试例程为了测试子进程不继承父进程的mmap。父进程打开一个文件sample.txt，然后将这个文件 通过mmap映射到`0x54321000`这个地址上，然后创建一个子进程child-inherit，子进程去写父进程的那个`0x54321000`这个地 址，这个应该会失败，导致子进程产生异常退出，父进程最终退出。父子进程退出时，都会清除自己的`mmap_list`上的项。
```
在SYS_MMAP处设置断点：
fd:3 addr:0x54321000
系统调用调用的是内核中的mmap()例程
mmap的文件大小为794bytes
这个例程和load()其实干的一些工作类似
load()是根据ELF文件格式导入，mmap()是直接将文件导入
两种导入其实都没有真正将文件内容导入到内存，而是增加struct thread结构体中的spt哈希表表项

后面果然会有fault_addr为0x54321000的page fault发生
这个page fault应该是由memcmp (actual, sample, strlen (sample))触发

然后父进程创建child-inherit子进程
子进程的memset ((char *) 0x54321000, 0, 4096)同样会触发fault_addr为0x54321000的page fault
但是由于没有继承父进程的mmap信息，这个地址根本不在子进程的地址空间，会导致子进程异常退出
子进程退出会清楚自己的mmap信息，但是mmap_list上的记录表项为0

子进程退出后，父进程也可以退出了，退出时也要清楚自己的mmap信息
父进程上有一项mmap信息，清楚mmap信息，以及mmap对应的spte信息
这里还有一个问题就是mmap对应的file到底怎么关闭
file信息在struct thread的open_file_list上也有
只能关闭一次，关闭两次会导致kernel panic，这里我选择仅仅在open_file_list上关闭
```
上面的调试记录，以及实验结果也说明了MMAP系统调用是可以正确工作的。
