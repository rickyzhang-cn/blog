<!---
title:: Pintos文件系统初探1
date:: 2015-04-11 21:08
categories:: 系统与网络
tags:: c, pintos, fs
-->
## `ide.c`和`block.c`
这两个文件是作为文件载体的硬件的设备驱动程序，是文件系统的硬件级底层支持。
````
/* Initialize the disk subsystem and detect disks. */
void
ide_init (void) 
{
    size_t chan_no;

    for (chan_no = 0; chan_no < CHANNEL_CNT; chan_no++)
    {
        struct channel *c = &channels[chan_no];
        int dev_no;

        /* Initialize channel. */
        snprintf (c->name, sizeof c->name, "ide%zu", chan_no);
        switch (chan_no) 
        {
            case 0:
                c->reg_base = 0x1f0;
                c->irq = 14 + 0x20;
                break;
            case 1:
                c->reg_base = 0x170;
                c->irq = 15 + 0x20;
                break;
            default:
                NOT_REACHED ();
        }
        lock_init (&c->lock);
        c->expecting_interrupt = false;
        sema_init (&c->completion_wait, 0);

        /* Initialize devices. */
        for (dev_no = 0; dev_no < 2; dev_no++)
        {
            struct ata_disk *d = &c->devices[dev_no];
            snprintf (d->name, sizeof d->name,
                    "hd%c", 'a' + chan_no * 2 + dev_no); 
            d->channel = c;
            d->dev_no = dev_no;
            d->is_ata = false;
        }

        /* Register interrupt handler. */
        intr_register_ext (c->irq, interrupt_handler, c->name);

        /* Reset hardware. */
        reset_channel (c);

        /* Distinguish ATA hard disks from other devices. */
        if (check_device_type (&c->devices[0]))
            check_device_type (&c->devices[1]);

        /* Read hard disk identity information. */
        for (dev_no = 0; dev_no < 2; dev_no++)
            if (c->devices[dev_no].is_ata)
                identify_ata_device (&c->devices[dev_no]);
    }
}
````
系统初始化时会调用ide_init()，该例程初始化disk subsystem并且探测disk的存在。最重要的是通过intr_register_ext()为每一个通道上的ide设备注册中断服务程序，然后identify_ata_device()例程中会通过block_register()例程注册block device。

````
/* Registers a new block device with the given NAME.  If
   EXTRA_INFO is non-null, it is printed as part of a user
   message.  The block device's SIZE in sectors and its TYPE must
   be provided, as well as the it operation functions OPS, which
   will be passed AUX in each function call. */
struct block *
block_register (const char *name, enum block_type type,
                const char *extra_info, block_sector_t size,
                const struct block_operations *ops, void *aux)；
上面是block_register()例程的一个原型说明。

block = block_register (d->name, BLOCK_RAW, extra_info, capacity,
                          &ide_operations, d);
partition_scan (block);
identify_ata_device()先对block_register()的调用，此时会terminal上会输出：
hda: 13,104 sectors (6 MB), model "BXHD00011", serial "Generic 1234"
identify_ata_device()然后对partition_scan()进行调用，就是这个ATA磁盘上的分区信息，会输出：
hda1: 186 sectors (93 kB), Pintos OS kernel (20)
hda2: 4,096 sectors (2 MB), Pintos file system (21)
hda3: 222 sectors (111 kB), Pintos scratch (22)
hda4: 8,192 sectors (4 MB), Pintos swap (23)

ide_init()运行完成后，运行locate_devices()例程，会输出：
filesys: using hda2
scratch: using hda3
swap: using hda4
/* Figure out what block devices to cast in the various Pintos roles. */
static void locate_block_devices (void);
````
有个问题需要阐述一下，就是关于磁盘读写例程的问题：
````
首先看一看注册ATA磁盘中的传入参数：
static struct block_operations ide_operations =
{
    ide_read,
    ide_write
};
也就是ide_read()和ide_write()是读写例程
/* Reads sector SEC_NO from disk D into BUFFER, which must have
   room for BLOCK_SECTOR_SIZE bytes.
   Internally synchronizes accesses to disks, so external
   per-disk locking is unneeded. */
static void
ide_read (void *d_, block_sector_t sec_no, void *buffer)
{
    struct ata_disk *d = d_;
    struct channel *c = d->channel;
    lock_acquire (&c->lock);
    select_sector (d, sec_no);
    issue_pio_command (c, CMD_READ_SECTOR_RETRY);
    sema_down (&c->completion_wait);
    if (!wait_while_busy (d))
        PANIC ("%s: disk read failed, sector=%"PRDSNu, d->name, sec_no);
    input_sector (c, buffer);
    lock_release (&c->lock);
}

/* Write sector SEC_NO to disk D from BUFFER, which must contain
   BLOCK_SECTOR_SIZE bytes.  Returns after the disk has
   acknowledged receiving the data.
   Internally synchronizes accesses to disks, so external
   per-disk locking is unneeded. */
static void
ide_write (void *d_, block_sector_t sec_no, const void *buffer)
{
    struct ata_disk *d = d_;
    struct channel *c = d->channel;
    lock_acquire (&c->lock);
    select_sector (d, sec_no);
    issue_pio_command (c, CMD_WRITE_SECTOR_RETRY);
    if (!wait_while_busy (d))
        PANIC ("%s: disk write failed, sector=%"PRDSNu, d->name, sec_no);
    output_sector (c, buffer);
    sema_down (&c->completion_wait);
    lock_release (&c->lock);
}
这里使用了一些同步机制，保证一致性，这个没有仔细去思考，也不需要多关注
真正实施读写的是input_sector()和out_sector()这两个函数，同时都是以sector为单位读写的
````
上面是硬件层的读写例程，真正提供给文件系统调用的例程都在block.c中。
````
/* Reads sector SECTOR from BLOCK into BUFFER, which must
   have room for BLOCK_SECTOR_SIZE bytes.
   Internally synchronizes accesses to block devices, so external
   per-block device locking is unneeded. */
void
block_read (struct block *block, block_sector_t sector, void *buffer)
{
    check_sector (block, sector);
    block->ops->read (block->aux, sector, buffer);
    block->read_cnt++;
}

/* Write sector SECTOR to BLOCK from BUFFER, which must contain
   BLOCK_SECTOR_SIZE bytes.  Returns after the block device has
   acknowledged receiving the data.
   Internally synchronizes accesses to block devices, so external
   per-block device locking is unneeded. */
void
block_write (struct block *block, block_sector_t sector, const void *buffer)
{
    check_sector (block, sector);
    ASSERT (block->type != BLOCK_FOREIGN);
    block->ops->write (block->aux, sector, buffer);
    block->write_cnt++;
}
block的读写例程最终还是通过调用注册的ide_read()和ide_write()来完成
通过block_sector_t这个uint32_t类型来表示哪一个sector被操作
````
## `filesys_init()`
在ATA磁盘硬件初始化完成之后，进行文件系统的初始化，文件系统的初始化由filesys_init()完成。
````
/* Initializes the file system module.
   If FORMAT is true, reformats the file system. */
void
filesys_init (bool format) 
{
    fs_device = block_get_role (BLOCK_FILESYS);
    if (fs_device == NULL)
        PANIC ("No file system device found, can't initialize file system.");

    inode_init ();
    free_map_init ();

    if (format) 
        do_format ();

    free_map_open ();
}
````
其中inode的初始化比较简单，就是初始化open_inodes这个链表。free_map的初始化是基于bitmap来进行的，通过bitmap来管理磁盘上sector的分配与释放。

关于文件系统上的一些layour信息见下面这个图：

<p><a href="http://www.rickyzhang.me/blog/wp-content/uploads/2015/04/filesys_layout.jpg"><img class="alignnone size-medium wp-image-740" src="http://www.rickyzhang.me/blog/wp-content/uploads/2015/04/filesys_layout-300x183.jpg" alt="filesys_layout" width="300" height="183" /></a></p>
文件系统的初始化工作按照上面的图进行的。

前面两个sector预留，分别用来记录freemap文件的inode信息和root directory文件的inode信息。freemap文件记录的就是整个磁盘上的sector的使用情况，占用4个sector，总共能记录`4*512*8`个sectors，因为bitmap是每一个bytes可以记录8个sectors，这个对于给file system分区大小为2MB已经足够。
