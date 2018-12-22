<!---
title:: Pintos实验阶段总结2
date:: 2015-01-12 21:08
categories:: 系统与网络
tags:: c, pintos, thread, schedule
-->

最近算是将Pintos实验的Project1完成得差不多，除了BSD4.4的MLFQ调度器暂且搁置外，其他优先级调度和Priority Donation的内容全部通过测试。Project2也慢慢开始了，以后有时间还是回来完成MLFQ的调度的。
## 优先级调度
目前的Pintos的调度方式是基于优先级调度，也就是优先级高的thread优先运行，而且Pintos的thread是可以抢占的。

1.`thread_create()`中会检查当前thread是否可以被抢占，如果可以，会触发thread切换，切换由当前thread调用`thread_yield()`主动放弃完成；

2.时钟中断到来时，一方面时间片用完会引发thread的切换，优先级相同的thread采用Ronud-Robin机制轮询调度，另一方面，会检查当前进程是否会被更高优先级抢占，也就是看全局`ready_list`中是否有优先级更高的thread；

注：因为有可能有些thread的priority在当前thread运行过程中产生了变化，不过可以在变化的地方添加`check_preemption()`应该就可以不必在时钟中断冗余添加这个判断了，如果添加了mlfq调度器的话，中断中就必须添加了。

3.对于sema信号量，等待在同一个sema上的thread，当`sema_up()`运行也是根据优先级将等待其上的thread投入到`ready_list`的，这也可能会引发抢占；

4.基于优先级的调度和同步机制在一起就会有问题，比如线程H、M、L的优先级依次递减，L获得了lock A，而H在wait lock A，这样处于`ready_list`中的M会优先于M先运行，这样H和M之间就发生了“优先级反转”的问题，需要去解决。

首先关于根据优先级选择thread运行主要和`ready_list`相关，插入新的thread进入到`ready_list`保证其逆序就OK，然后`ready_list`中的第一个thread就是优先级最高的thread，而且由于`ready_list`是系统全局资源，可以再无论什么情况下，当`ready_list`中的thread的优先级发生变化可以对其重新排序。

至于sema的waiter和`ready_list`基本一样，wait在某一sema A上的thread都在A的waiter双链表上，处于`THREAD_BLOCKED`状态，当`sema_up()`也是根据优先级投入到`ready_list`上，但是sema的waiter链表上的thread并不是有序的，而是乱序的，通过`list_min()/list_max()`找出优先级最高的thread，这是因为`ready_list`的重排序更加方面，如果wait某一sema的thread的优先级发生变化，重排序不是很好办，所以我就选择了不是有序插入，取出优先级最高的就需要遍历waiter这个链表了。

## Priority Donation
前面优先级调度中也提到了“优先级反转”的问题，这个问题的解决通过Priority Donation的策略来解决，Pintos中也只解决lock的这个问题。简单来说，就是当A wait B时，如果A的优先级高于B就将B的优先级设置为A的优先级，在A释放lock时恢复其原来的优先级，这是一个最基本的情况。

复杂点的情况有两种，一种是当A wait B时，B在wait C，C可能还在wait D，也就是wait构成了一个chain，这种情况怎么处理，这就需要一个循环，对chain上的每个thread的优先级设置为A的优先级和本身优先级的较大值。另一种情况是当A拥有多个lock时，在释放其中一个lock时，其优先级的设置问题，这时应该其优先级为自身优先级，和wait在其他lock上的优先级的最大值。

第一种情况的代码：
````
void sema_down (struct semaphore *sema) 
{
    enum intr_level old_level;

    ASSERT (sema != NULL);
    ASSERT (!intr_context ());

    old_level = intr_disable ();
    while (sema->value == 0) 
    {
        //list_push_back (&sema->waiters, &thread_current ()->elem);
        list_insert_ordered(&sema->waiters,&thread_current()->elem,less_func,"p");
        thread_current()->blocking_lock=sema->containing_lock;

        if((sema->containing_lock != NULL) && !list_empty(&(sema->waiters)))
        {
            struct list_elem *waiters_top=list_begin(&(sema->waiters));
            struct thread *waiters_top_thread=list_entry(waiters_top,struct thread,elem);
            struct thread *lock_holder=(waiters_top_thread->blocking_lock)->holder;
            struct semaphore *s=sema;

            while(lock_holder->blocking_lock != NULL)
            {
                waiters_top=list_begin(&(s->waiters));
                waiters_top_thread=list_entry(waiters_top,struct thread,elem);
                lock_holder=(waiters_top_thread->blocking_lock)->holder;

                thread_set_ready_priority(lock_holder,waiters_top_thread->priority);

                if((lock_holder->blocking_lock) != NULL)
                {
                    s=&((lock_holder->blocking_lock)->semaphore);
                }
            }
            thread_set_ready_priority(lock_holder,waiters_top_thread->priority);
        }
        thread_block ();
    }
    sema->value--;
    intr_set_level (old_level);
}

void thread_set_ready_priority(struct thread *t,int new_priority)
{
    ASSERT(new_priority >= PRI_MIN && new_priority <= PRI_MAX);
    if(new_priority > (t->priority))
    {
        t->under_donation=true;
        t->priority=new_priority;

        //list_sort(&ready_list,&less_func,"p");
        ready_list_reorder();
    }
}
````
第二种情况的代码：
````
void sema_up (struct semaphore *sema) 
{
    enum intr_level old_level;

    ASSERT (sema != NULL);

    old_level = intr_disable ();
    if (!list_empty (&sema->waiters))
    {
        // thread_unblock (list_entry (list_pop_front (&sema->waiters),
        //                           struct thread, elem));
        struct list_elem *top_elem=list_pop_front(&sema->waiters);
        struct thread *t=list_entry(top_elem,struct thread,elem);

        if(sema->containing_lock!=NULL)
        {
            struct thread *lock_holder=(sema->containing_lock)->holder;
            //struct thread *lock_holder=thread_current();

            thread_set_true_priority(lock_holder);

            //msg("in sema_up,lock_holder,name=%s priority=%d\n",lock_holder->name,lock_holder->priority);

            t->blocking_lock=NULL;
            sema->containing_lock=NULL;
        }

        thread_unblock(t);
    }
    sema->value++;
    intr_set_level (old_level);

    if(check_preemption())
    {
        thread_yield();
    }
}


void thread_set_true_priority(struct thread *t)
{
    t->under_donation=false;

    struct list_elem *e;
    struct list_elem *te;
    struct thread *top_thread;
    struct lock *l;
    int max_priority=t->actual_priority;

    if(list_size(&(t->locks_list)) > 0)
    {
        for(e=list_begin(&(t->locks_list));e!=list_end(&(t->locks_list));e=list_next(e))
        {
            l=list_entry(e,struct lock,locks_elem);
            if(list_empty(&((l->semaphore).waiters)))
                continue;
            else
            {
                te=list_begin(&((l->semaphore).waiters));
                top_thread=list_entry(te,struct thread,elem);
                if(top_thread->priority > max_priority)
                    max_priority=top_thread->priority;
            }
        }
    }
    t->priority=max_priority;
    //msg("in thread_set_true_priority,t,name=%s priority=%d\n",t->name,t->priority);
    //list_sort(&ready_list,less_func,"p");
    ready_list_reorder();
}
````
这里的代码其实没有考虑到加入一个thread等待在多个lock上，也就是这个实现限制线程只能可以获得多个lock，却只能wait一个lock，这个下次可以改进。
## Tips
1.关于同步的问题，Pintos中需要关中断保持一致性的全局资源是中断服务程序可能修改的资源，比如`ready_list`，在引入mlfq调度器后，时钟中断服务程序中需要计算`ready_list`中thread的优先级。其他中断服务程序不会修改的资源，可以使用lock或者sema保证一致性。

2.多使用gdb调试，不要想当然，用gdb调试观察现象，然后分析，很多问题没有自己想象中的复杂。

