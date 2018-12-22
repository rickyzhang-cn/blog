<!---
title:: Pintos实验环境搭建
date:: 2014-11-26 21:08
categories:: 系统与网络
tags:: c, pintos, thread
-->

最近决定系统学习一下OS这门课，决定跟着UC Berkeley的CS162和MIT的6.828这两门课程来学习，主要是进行一些实践，CS162中实验是基于Stanford开发的用于教学的操作系统Pintos。

首先是去下载一份pintos的源码包，下载地址：[Pintos](http://www.stanford.edu/class/cs140/projects/pintos/pintos.tar.gz)

在开始之前，关于Pintos的一个简单介绍你可以看其官方说明文档：[Pintos](http://web.stanford.edu/class/cs140/projects/pintos/pintos_1.html)

我选择32bit的Ubuntu上使用Bochs来运行调试Pintos，主要是因为我使用过Bochs，而且觉得对于这种简单OS的调试还是很适合的。

在Pintos的源码包中的misc/目录中有一个安装带调试功能bochs的脚本bochs-2.2.6-build.sh，可以自己看一看，就用这个脚本进行安装就可以

脚本中使用的bochs源码版本是2.2.6，你需要去下载一个2.2.6的bochs源码包。然后运行下面的命令：
````
#下面这个包如果你的电脑上有就不需要安装
sudo apt-get install gcc binutils perl make gdb qemu g++ libwxbase2.8-0 libwxgtk2.8-dev libwxgtk2.8-dbg libxmu-dev libxmuu-dev libncurses5-dev
#安装bochs2.2.6
cd /path/to/pintos/src/misc
env SRCDIR=/path/to/bochs-2.2.6.tar.gz PINTOSDIR=/path/to/pintos DSTDIR=/usr/local sh ./bochs-2.2.6-build.sh
#将pintos相关的一些辅助工具放入系统$PATH变量中
cd /path/to/root/of/pintos/src/utils
sudo make
sudo cp backtrace pintos pintos-gdb pintos-mkdisk Pintos.pm squish-pty /usr/local/bin
````

还有一个就是需要修改pintos-gdb这个perl脚本中GDBMACROS变量为源码包util/目录中gdb-macros文件的位置。

Pintos的基本环境基本就搭建好了，下面就是使用和添加一些代码调试了。
