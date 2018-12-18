<!---
title:: Pintos实验阶段总结1
date:: 2014-12-23 21:08
categories:: 系统与网络
tags:: c, pintos, thread
-->

自从决定开始通过Pintos实验重新学习操作系统以来，已经快一个月了，这一个月，虽然实验室项目没有投入很多精力，但是经历了几门考试，所以对Pintos实验的投入时间不是很多，但是自己终究还是一直没有忘记这个事，还是完成了一些任务了。

## 对thread的一点理解
在pintos中，每个thread结构体都占用4kb的内存page，关于thread的一些信息和stack都在这个page上，有数据结构相关的elem字段，也有tid、status、priority这样thread本身属性字段，这个结构体在后面估计需要不停扩充。

一个thread的建立，静态的就有代码段，静态数据区等，运行时环境就有stack和一些CPU寄存器值的保存副本，以及向申请的一些资源，比如内存等。执行的代码运行流由CPU EIP指示，Stack的栈顶由CPU的ESP指示。一个thread就是一个代码执行流，其中function call通过自己的栈来完成。多个thread，必然会有切换，进程的切换，就是代码执行流的切换，必然会有EIP和ESP的切换，很多切换相关的信息在thread自己的stack中，这部分在pintos中看到有一些代码，不是特别明白，在thread_create()时，就会预先将struct kernel_thread_frame，struct switch_entry_frame和switch_thread_frame压栈，这部分我觉得可能是和switch.S中的switch_entry()和switch_threads()函数有关。

关于pintos中thread的status的状态转移图，自己是仔细琢磨了一番，这个很值得去思考，哪些函数，哪些因素会导致thread的状态发生变化。在pintos中，可以通过thread_yield()主动放弃CPU，也可以因为时间片到了，在中断中被动放弃CPU，不过导致thread放弃CPU的函数都是thread_yield()。

## 中断处理
pintos中的中断处理，目前还没有怎么去仔细研究，什么IDT表的建立都没有深入研究。在pintos中，所有中断都对应一个中断处理函数intr_handler()这个函数，但是会传进来一个参数struct intr_frame *frame，通过这个参数可以根据中断向量号调用中断处理函数，然后调用玩中断处理函数后，还会有一些后续处理，比如判断被中断的thread是否应该被动放弃CPU，切换thread就是在这个函数中处理，以及将8259A的相应端口的处理也是在这里进行。


## timer_sleep()的改进
原始的pintos中timer_sleep()是忙等待，这个显然是不科学的，需要修改。主要的思路就是添加一个和ready_list作用类似的sleeping_list，在thread结构体中添加sleeping_ticks字段。将所有调用timer_sleep()函数的thread根据wakeup时间的由小到大插入到sleeping_list中。然后在中断处理中判断是否该唤醒，将睡眠时间到的函数根据优先级由大到小插入到ready_list中。

在pintos中有list.h和list.c库，需要的链表操作例程基本都有，基本够用，但是一定要用好，要模拟链表变化，自己的bug基本都是出在链表操作不正确上。

## 基于优先级的调度
这部分暂时只是完成了部分，主要是对ready_list的一些操作，ready_list是一个根据priority字段discending order排序的链表，只要是ready_list中添加了新的thread，都需要去判断是否需要抢占当前进程，还有优先级反转的问题，需要优先级donate来实现，这些都有待进一步工作，目前只完成了简单的抢占，通过了一部分测试。

## 其他
自己学习pintos以来，发现国外计算机教育真心很好，整个pintos工程，不仅仅可以让你学习到操作系统有关的东西，还可以学习到如何去做一个实际项目的东西，整个工程的组织很赞，值得花时间去研究，以及对调试的支持，还有文档，都很高水准。

Pintos一系列实验值得自己花时间好好研究，慢慢研究。一定要坚持，遇到困难不要退缩，总会找到解决方法的，发呆不理会是最差的选择。自己对整个工程也算是很了解，以后coding和thinking的时间多投入，reading的话只是在一些碎片时间去辅助理解的，源码面前，了无秘密，也要合理安排时间。

整个实验过程也通过github来记录：[Pintos](https://github.com/rickyzhang-cn/Pintos)
