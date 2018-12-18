<!---
title:: Pintos实验阶段总结3
date:: 2015-02-02 21:08
categories:: 系统与网络
tags:: c, pintos, thread
-->

最近这段时间还算时间比较多，在把相关代码和文档大概看了看之后，开始着手完成Project2。Project2最主要任务主要有2个：支持ELF格式的可执行User Program的载入运行和实现必要的一些system call。

关于User Program的载入运行在前面的博文中已经有一些说明，也就是通过申请内存页给User Program建立好运行时环境，通过一个中断退出例程让eip指向User Program的执行入口，esp指向load()例程建立好的栈的栈顶。
<h2>Kernel Stack VS User Stack</h2>
由于Pintos支持User Program的运行，一个user process对应一个kernel thread，这样一个user process拥有2个栈，一个是在用户空间user stack，一个是在内核空间的kernel stack。这就涉及到kernel stack和user stack的切换问题，只要是用户空间和内核空间切换时就会触发stack的切换，这里涉及到tss_update()这个例程。
<pre class="brush:cpp">void
tss_update (void) 
{
  ASSERT (tss != NULL);
  tss-&gt;esp0 = (uint8_t *) thread_current () + PGSIZE;
}
</pre>
关于tss的相关细节还需要去参考Intel的官方文档，Pintos其实弱化了X86中tss机制的作用，Pintos中建立一个和tss有关的段选择子，具体细节的一些机制要去研究哈Intel官方文档。我没有仔细去研究，但是我觉得大概是这个样子，在tss结构体中时时刻刻记录了当前user process对应的内核栈的信息，也就是用户进程的切换必然也会有tss的切换。

Pintos处于下列3个状态之一：A.进程上下文，用户空间；B.进程上下文，内核空间；C.中断上下文，内核空间。当A进入B时，一般是用户程序调用system call，这时cpu就可以通过tss知道内核栈的栈顶，完成栈的切换，A进入C时也这样完成stack切换。B或者C到A，是因为A到B或者C是，切换到kernel stack后会将用户的一些上下文(context)信息压栈到kernel stack上，返回时pop出来即可。B和C之间的切换不涉及stack的切换，直接都是使用同一个kernel stack。

在load()例程中，会在load ELF文件到内存时会先调用tss_update()这个例程，这个例程主要是保存该User Program对应的内核thread的kernel stack的栈顶，而且这时kernel stack为空。这样做最主要的目的就是当User Program被中断服务程序中断时可以找到其内核栈。在thread切换例程的最后，也会调用tss_update()，完成thread切换时，tss中stack的切换。
<h2>系统调用的实现</h2>
系统调用是用户程序和内核的接口，用户程序通过系统调用进入内核空间，使用内核提供的服务。Pintos中系统调用时通过中断来实现从用户空间进入内核空间，中断号为0x30.

在用户空间中，根据下面的GCC编译参数，可以看到关于系统调用的API函数的实现都编译进静态链接的C库libc.a中，最后编译到用户程序中。
<pre class="brush:plain">gcc  -Wl,--build-id=none -nostdlib -static -Wl,-T,../../lib/user/user.lds
 tests/userprog/bad-write.o tests/main.o tests/lib.o lib/user/entry.o libc.a 
-o tests/userprog/bad-write</pre>
用户空间中，关于系统调用最主要的代码片段参见下面的一段代码，这段代码是通过内联汇编来实现，主要就是先将用户参数按照从右到左压栈，这里的栈是user stack，不是kernel stack。压栈的有系统调用号和系统调用的传递参数，这些最终在内核空间中可以通过kernel stack中保存的用户空间上下文的信息可以找到。中断返回时修改esp，舍弃和系统调用压栈的参数。
<pre class="brush:cpp">/* Invokes syscall NUMBER, passing no arguments, and returns the
   return value as an `int'. */
#define syscall0(NUMBER)                                        \
        ({                                                      \
          int retval;                                           \
          asm volatile                                          \
            ("pushl %[number]; int $0x30; addl $4, %%esp"       \
               : "=a" (retval)                                  \
               : [number] "i" (NUMBER)                          \
               : "memory");                                     \
          retval;                                               \
        })

/* Invokes syscall NUMBER, passing argument ARG0, and returns the
   return value as an `int'. */
#define syscall1(NUMBER, ARG0)                                           \
        ({                                                               \
          int retval;                                                    \
          asm volatile                                                   \
            ("pushl %[arg0]; pushl %[number]; int $0x30; addl $8, %%esp" \
               : "=a" (retval)                                           \
               : [number] "i" (NUMBER),                                  \
                 [arg0] "g" (ARG0)                                       \
               : "memory");                                              \
          retval;                                                        \
        })

/* Invokes syscall NUMBER, passing arguments ARG0 and ARG1, and
   returns the return value as an `int'. */
#define syscall2(NUMBER, ARG0, ARG1)                            \
        ({                                                      \
          int retval;                                           \
          asm volatile                                          \
            ("pushl %[arg1]; pushl %[arg0]; "                   \
             "pushl %[number]; int $0x30; addl $12, %%esp"      \
               : "=a" (retval)                                  \
               : [number] "i" (NUMBER),                         \
                 [arg0] "g" (ARG0),                             \
                 [arg1] "g" (ARG1)                              \
               : "memory");                                     \
          retval;                                               \
        })

/* Invokes syscall NUMBER, passing arguments ARG0, ARG1, and
   ARG2, and returns the return value as an `int'. */
#define syscall3(NUMBER, ARG0, ARG1, ARG2)                      \
        ({                                                      \
          int retval;                                           \
          asm volatile                                          \
            ("pushl %[arg2]; pushl %[arg1]; pushl %[arg0]; "    \
             "pushl %[number]; int $0x30; addl $16, %%esp"      \
               : "=a" (retval)                                  \
               : [number] "i" (NUMBER),                         \
                 [arg0] "g" (ARG0),                             \
                 [arg1] "g" (ARG1),                             \
                 [arg2] "g" (ARG2)                              \
               : "memory");                                     \
          retval;                                               \
        })

#if 0
/* Invokes syscall NUMBER, passing arguments ARG0, ARG1, and
   ARG2, and returns the return value as an `int'. */
#define syscall3(NUMBER, ARG0, ARG1, ARG2)                      \
        ({                                                      \
          int retval;                                           \
          asm volatile                                          \
            ("movl %[arg2],-4(%esp); movl %[arg1],-8(%esp); movl %[arg0],-12(%esp); "    \
             "movl %[number],-16(%esp); subl $16, %%esp; int $0x30; addl $16, %%esp"      \
               : "=a" (retval)                                  \
               : [number] "i" (NUMBER),                         \
                 [arg0] "g" (ARG0),                             \
                 [arg1] "g" (ARG1),                             \
                 [arg2] "g" (ARG2)                              \
               : "memory");                                     \
          retval;                                               \
        })
#endif

#if 0
/* Invokes syscall NUMBER, passing arguments ARG0, ARG1, and
   ARG2, and returns the return value as an `int'. */
#define syscall3(NUMBER, ARG0, ARG1, ARG2)                      \
        ({                                                      \
          int retval;                                           \
          asm volatile                                          \
        ("pushl 0xc(%esp);pushl 0xc(%esp); pushl 0xc(%esp); "   \
             "pushl %[number]; int $0x30; addl $16, %%esp"      \
               : "=a" (retval)                                  \
               : [number] "i" (NUMBER),                         \
                 [arg0] "g" (ARG0),                             \
                 [arg1] "g" (ARG1),                             \
                 [arg2] "g" (ARG2)                              \
               : "memory");                                     \
          retval;                                               \
        })
#endif
</pre>
上面都是用户空间的相关说明，对于用户空间来说，问题还比较简单。下面说明内核空间的处理，进入内核空间后，处于中断上下文，中断服务例程开始执行。在内核空间，编程者必须小心对待从用户空间传进来的指针以及参数，所以必须对这些参数进行检查，防止让内核panic。

系统调用中断服务程序就是通过kernel stack中记录的用户空间上下文的esp指针指向的user stack的栈顶来获得系统调用的参数值，然后根据这些参数进行相应处理，以及一些必要的错误处理。

在实现系统调用中，我发现一个很不正常的现象，应该算是Pintos自身的一个bug吧，或者是在有些编译器下会存在的bug吧。这个bug最主要的就是，就举write这个系统调用为例吧，下面是通过objdump获得的libc.a中关于write系统调用的反汇编代码：
<pre class="brush:plain">000000ca &lt;write&gt;:
  ca:   ff 74 24 0c             pushl  0xc(%esp)
  ce:   ff 74 24 08             pushl  0x8(%esp)
  d2:   ff 74 24 04             pushl  0x4(%esp)
  d6:   6a 09                   push   $0x9
  d8:   cd 30                   int    $0x30
  da:   83 c4 10                add    $0x10,%esp
  dd:   c3                      ret
</pre>
一个很严重的问题就是pushl是esp也在变，下一次pushl后就不是我们真正想压栈的值了，所以导致刚开始实现这些系统调用时都无法正常工作。开始我觉比较优美点的解决方法是，不通过pushl来压栈，而是通过movl来压栈，最后addl esp来修改栈顶，但是自己汇编不是很好，尝试修改了几次都编译不通过，也就只有选择tricky点的方式了。其实在这些参数压栈前，栈中其实就有这些参数的值，除了系统调用号之外。从反汇编中也可以看出，在esp不变的情况下，0xc(%esp)就是write的第三个参数，0x8(%esp)和0x4(%esp)就分别是第二个和第一个参数，这些参数之所以在user stack中是基于C语言的function call convention，也就是函数调用时，由函数调用者将参数从右到左压栈，jmp时还会将当前的eip压栈。所以进入系统调用后用tricky点的方式获得这些参数即可，就是在user stack中去找到正确的参数即可，反正user stack中这些信息都有，想要获得它们并不难。