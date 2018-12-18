<!---
title:: Pintos中User Program运行机制的探索
date:: 2015-01-23 21:08
categories:: 系统与网络
tags:: c, pintos, thread
-->

在Pintos中一个用户态的thread对应着一个内核态的thread，在Pintos中，这个特殊的thread的运行入口函数均是同一个函数，也就是start_process()，而且传递了一个参数给start_process()。这个过程由process_execute()执行。
<pre class="brush:cpp">/* A thread function that loads a user process and starts it
   running. */
static void
start_process (void *file_name_)
{
  char *file_name = file_name_;
  struct intr_frame if_;
  bool success;

  /* Initialize interrupt frame and load executable. */
  memset (&amp;if_, 0, sizeof if_);
  if_.gs = if_.fs = if_.es = if_.ds = if_.ss = SEL_UDSEG;
  if_.cs = SEL_UCSEG;
  if_.eflags = FLAG_IF | FLAG_MBS;

  char *rz_fn_copy;
  char *exec_file_name;
  char *saved_ptr;
  rz_fn_copy=palloc_get_page(0);
  strlcpy(rz_fn_copy,file_name,strlen(file_name)+1);
  exec_file_name=strtok_r(file_name," ",&amp;saved_ptr);

  printf("file name:%s\n",exec_file_name);

  success = load (exec_file_name, &amp;if_.eip, &amp;if_.esp);

  if(success)
  {
        struct thread *cur=thread_current();
        push_args(&amp;if_.esp,rz_fn_copy);
  }

  /* If load failed, quit. */
  palloc_free_page (file_name);

  palloc_free_page(rz_fn_copy);

  if (!success)
    thread_exit ();

  /* Start the user process by simulating a return from an
     interrupt, implemented by intr_exit (in
     threads/intr-stubs.S).  Because intr_exit takes all of its
     arguments on the stack in the form of a `struct intr_frame',
     we just point the stack pointer (%esp) to our stack frame
     and jump to it. */
  asm volatile ("movl %0, %%esp; jmp intr_exit" : : "g" (&amp;if_) : "memory");
  NOT_REACHED ();
}
</pre>
上面是start_process()函数例程的源码。这个函数中完成的最主要的任务有：
<ul>
	<li>将ELF文件格式的可执行文件从磁盘上load到内存中，这个过程由load()完成</li>
	<li>为用户thread创建一个运行时需要的stack，这个由load()函数例程中的setup_stack()完成</li>
	<li>将interrupt frame中的各个字段设置正确，特别是eip和esp这两个字段，eip指向的是ELF可执行文件的入口，esp指向的是用户态thread申请的用户栈的栈顶</li>
</ul>
对于用户态thread来说，拥有两个stack，一个是内核态的stack，一个是用户态的stack。对于内核态的stack，其大小为一个page除去struct thread结构体占用的部分；对于用户态stack，其大小也为一个page，也就是4KB，栈顶为0xc0000000，也就是3GB的地方。

对于用户态的thread来说，其内存分布为：

<a href="http://www.rickyzhang.me/blog/wp-content/uploads/2015/01/捕获.jpg"><img class="alignnone size-medium wp-image-716" src="http://www.rickyzhang.me/blog/wp-content/uploads/2015/01/捕获-280x300.jpg" alt="memory_map" width="280" height="300" /></a>

在start_process()中最重要的是最后一个inline assembly语句：

asm volatile ("movl %0, %%esp; jmp intr_exit" : : "g" (&amp;if_) : "memory");

对于inline assembly，其一般格式为：
<pre class="brush:plain"> asm ( assembler template 
           : output operands                  /* optional */
           : input operands                   /* optional */
           : list of clobbered registers      /* optional */
           );
</pre>
在上面的语句先将栈设置为自己设置好的interrupt frame的起始处，然后执行一个中断返回的例程，这个例程源码为：
<pre class="brush:plain">/* Interrupt exit.

   Restores the caller's registers, discards extra data on the
   stack, and returns to the caller.

   This is a separate function because it is called directly when
   we launch a new user process (see start_process() in
   userprog/process.c). */
.globl intr_exit
.func intr_exit
intr_exit:
        /* Restore caller's registers. */
        popal
        popl %gs
        popl %fs
        popl %es
        popl %ds

        /* Discard `struct intr_frame' vec_no, error_code,
           frame_pointer members. */
        addl $12, %esp

        /* Return to caller. */
        iret
.endfunc
</pre>
说明该例程的作用时，有必要给出interrupt frame的结构体声明：
<pre class="brush:cpp">/* Interrupt stack frame. */
struct intr_frame
  {
    /* Pushed by intr_entry in intr-stubs.S.
       These are the interrupted task's saved registers. */
    uint32_t edi;               /* Saved EDI. */
    uint32_t esi;               /* Saved ESI. */
    uint32_t ebp;               /* Saved EBP. */
    uint32_t esp_dummy;         /* Not used. */
    uint32_t ebx;               /* Saved EBX. */
    uint32_t edx;               /* Saved EDX. */
    uint32_t ecx;               /* Saved ECX. */
    uint32_t eax;               /* Saved EAX. */
    uint16_t gs, :16;           /* Saved GS segment register. */
    uint16_t fs, :16;           /* Saved FS segment register. */
    uint16_t es, :16;           /* Saved ES segment register. */
    uint16_t ds, :16;           /* Saved DS segment register. */

    /* Pushed by intrNN_stub in intr-stubs.S. */
    uint32_t vec_no;            /* Interrupt vector number. */

    /* Sometimes pushed by the CPU,
       otherwise for consistency pushed as 0 by intrNN_stub.
       The CPU puts it just under `eip', but we move it here. */
    uint32_t error_code;        /* Error code. */

    /* Pushed by intrNN_stub in intr-stubs.S.
       This frame pointer eases interpretation of backtraces. */
    void *frame_pointer;        /* Saved EBP (frame pointer). */

    /* Pushed by the CPU.
       These are the interrupted task's saved registers. */
    void (*eip) (void);         /* Next instruction to execute. */
    uint16_t cs, :16;           /* Code segment for eip. */
    uint32_t eflags;            /* Saved CPU flags. */
    void *esp;                  /* Saved stack pointer. */
    uint16_t ss, :16;           /* Data segment for esp. */
  };
</pre>
上面的intr_exit()例程就是进入到用户态，通过iret指令将esp指向申请的内核栈的栈顶，同时eip指向elf可指向文件的执行入口，我们可以通过objdump看到执行入口_start.
<pre class="brush:plain">08048709 &lt;_start&gt;:
 8048709:       83 ec 1c                sub    $0x1c,%esp
 804870c:       8b 44 24 24             mov    0x24(%esp),%eax
 8048710:       89 44 24 04             mov    %eax,0x4(%esp)
 8048714:       8b 44 24 20             mov    0x20(%esp),%eax
 8048718:       89 04 24                mov    %eax,(%esp)
 804871b:       e8 80 f9 ff ff          call   80480a0 &lt;main&gt;
 8048720:       89 04 24                mov    %eax,(%esp)
 8048723:       e8 6d 1a 00 00          call   804a195 &lt;exit&gt;
</pre>
通过设置interrupt frame的eip字段为elf文件执行入口，esp字段为申请的用户stack的栈顶，通过一个中断范围，进入到用户态，执行用户程序。有一个问题忽略了，那就是_start函数在调用main函数，也就是用户编写的代码时，有一个传argc和argv这两个参数的过程，所以我们在进入执行入口前，应该先为main函数设置好参数，做好给main函数参数传递的工作。

在用户态stack申请好后，就应该进行参数传递的准备工作，工作也就是将传递进来的参数压栈，然后将这些参数的地址压栈，最后将argv和argc以及return address依次压栈，最终的效果类似于以下：

<a href="http://www.rickyzhang.me/blog/wp-content/uploads/2015/01/args_stack.jpg"><img class="alignnone size-medium wp-image-717" src="http://www.rickyzhang.me/blog/wp-content/uploads/2015/01/args_stack-300x189.jpg" alt="args_stack" width="300" height="189" /></a>

在这个例子中，最终的stack pointer的值将是0xbfffffcc.
