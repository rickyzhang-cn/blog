<!---
title:: Pintos文件系统初探2
date:: 2015-04-12 21:08
categories:: 系统与网络
tags:: c, pintos, fs
-->

在ATA磁盘和文件系统均初始化完成后，Pintos就可以在磁盘上进行I/O操作了。
## `inode.c`
```
/* In-memory inode. */
struct inode 
{
    struct list_elem elem;              /* Element in inode list. */
    block_sector_t sector;              /* Sector number of disk location. */
    int open_cnt;                       /* Number of openers. */
    bool removed;                       /* True if deleted, false otherwise. */
    int deny_write_cnt;                 /* 0: writes ok, >;0: deny writes. */
    struct inode_disk data;             /* Inode content. */
};

/* On-disk inode.
   Must be exactly BLOCK_SECTOR_SIZE bytes long. */
struct inode_disk
{
    block_sector_t start;               /* First data sector. */
    off_t length;                       /* File size in bytes. */
    unsigned magic;                     /* Magic number. */
    uint32_t unused[125];               /* Not used. */
};
```
其中inode是In-memory inode，`inode_disk`是On-disk inode。`inode_disk`这个结构体正好是一个sector 512bytes的大小，最终会写到磁盘上的，`inode_disk`上记录的是一个磁盘文件的sector起点start和文件长度length。

Pintos的原始文件系统中只有根目录，没有多级目录。Root Directory也是一个`inode_disk`，Root Directory占用一个sector，里面存储的是文件的entry，记录文件`inode_disk`所在的sector，然后改sector记录的就是文件所在的sector起点start和占用的连续的sector个数length。
```
/* Initializes an inode with LENGTH bytes of data and
 writes the new inode to sector SECTOR on the file system
 device.
 Returns true if successful.
 Returns false if memory or disk allocation fails. */
bool inode_create (block_sector_t sector, off_t length);
inode_create()中传入的sector是之前通过free_map分配得到的sector号，length是文件的长度
inode_create()就是根据length确定需要多少个连续的sector，然后调用free_map_allocate()分配sector，记录sector起始start
最终将disk_inode这个记录start和length的结构体写到磁盘上

/* Reads an inode from SECTOR
 and returns a `struct inode' that contains it.
 Returns a null pointer if memory allocation fails. */
struct inode * inode_open (block_sector_t sector);
inode_open()首先根据sector在open_inodes链表中查找该sector对应的sector是否已经存在
如果存在，那么reopen该inode，如果不存在，使用malloc申请一个inode结构体，初始化该inode
每一个inode对应一个disk_inode，该inode的sector字段就是disk_inode所在的sector号

/* Reads SIZE bytes from INODE into BUFFER, starting at position OFFSET.
 Returns the number of bytes actually read, which may be less
 than SIZE if an error occurs or end of file is reached. */
off_t inode_read_at (struct inode *inode, void *buffer_, off_t size, off_t offset);
/* Writes SIZE bytes from BUFFER into INODE, starting at OFFSET.
 Returns the number of bytes actually written, which may be
 less than SIZE if end of file is reached or an error occurs.
 (Normally a write at end of file would extend the inode, but
 growth is not yet implemented.) */
off_t inode_write_at (struct inode *inode, const void *buffer_, off_t size,off_t offset);
这两个例程是file_read()和file_write()调用的读写例程，其中的offset值时文件的offset
```
## `file.c`
```
/* An open file. */
struct file 
{
    struct inode *inode;        /* File's inode. */
    off_t pos;                  /* Current position. */
    bool deny_write;            /* Has file_deny_write() been called? */

    int fd;
    struct list_elem open_file_elem;
};
```
file结构体构建在inode之上，添加了一些字段支持file的一些操作。
```
/* Opens a file for the given INODE, of which it takes ownership,
 and returns the new file. Returns a null pointer if an
 allocation fails or if INODE is null. */
struct file * file_open (struct inode *inode);
file_open()是根据传入的inode指针申请一个file结构体，初始化这个结构体
file_open()被上层的filesys_open()例程调用，filesys_open()传入参数为文件名

/* Reads SIZE bytes from FILE into BUFFER,
 starting at the file's current position.
 Returns the number of bytes actually read,
 which may be less than SIZE if end of file is reached.
 Advances FILE's position by the number of bytes read. */
off_t
file_read (struct file *file, void *buffer, off_t size) 
{
    off_t bytes_read = inode_read_at (file->;inode, buffer, size, file->;pos);
    file->;pos += bytes_read;
    return bytes_read;
}
/* Writes SIZE bytes from BUFFER into FILE,
 starting at the file's current position.
 Returns the number of bytes actually written,
 which may be less than SIZE if end of file is reached.
 (Normally we'd grow the file in that case, but file growth is
 not yet implemented.)
 Advances FILE's position by the number of bytes read. */
off_t
file_write (struct file *file, const void *buffer, off_t size) 
{
    off_t bytes_written = inode_write_at (file->;inode, buffer, size, file->;pos);
    file->;pos += bytes_written;
    return bytes_written;
}
file的读写例程，系统调用中文件的读写就是调用这里的读写例程
```
## `filesys.c`
一直觉得filesys.c这个文件中的例程比较杂乱，有系统初始化时调用的`filesys_init()`例程，也有被系统调用机制中调用的`filesys_create/open/remove`例程。

这个文件中的例程基本上是`directory.c`最大的使用者，因为`filesys.c`中的函数创建文件，打开文件，删除文件，这个会经常操作Root Directory。
## `fsutil.c`
`fsutil.c`相当于是应用层的一些例程，这些例程最主要在系统初始化有使用，将scratch磁盘上打包文件解压出来，将文件写入到文件系统上，准备好用户的测试程序。

<a href="http://www.rickyzhang.me/blog/wp-content/uploads/2015/04/filesys_used.jpg"><img class="alignnone size-medium wp-image-748" src="http://www.rickyzhang.me/blog/wp-content/uploads/2015/04/filesys_used-300x185.jpg" alt="filesys_used" width="300" height="185" /></a>

scratch disk上的文件是由perl脚本写入进去的，是一个压缩文件，系统在启动后会通过`fsutil_extract()`例程将文件解压然后写入到filesystem disk上去。

