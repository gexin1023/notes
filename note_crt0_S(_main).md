# crt0,S(_main)代码分析
---
## 1. 设置sp寄存器地址
``` C
//设置SP栈指针
#if defined(CONFIG_SPL_BUILD) && defined(CONFIG_SPL_STACK)
	ldr	sp, =(CONFIG_SPL_STACK)
#else
	ldr	sp, =(CONFIG_SYS_INIT_SP_ADDR)
#endif

//设置地址八位对齐
#if defined(CONFIG_CPU_V7M)	/* v7M forbids using SP as BIC destination*/
	mov	r3, sp			   
	bic	r3, r3, #7		   //* 后三位清零相当于堆栈地址八位对齐
	mov	sp, r3				 
#else
	bic	sp, sp, #7	/* 8-byte alignment for ABI compliance */	
#endif
```
## 2. 在栈中为全局变量gd分配空间

```
// r0寄存器传递函数的参数
mov	r0, sp
bl	board_init_f_alloc_reserve	//在栈中为全局数据分配空间
mov	sp, r0		
//函数调用后返回值在r0中，将其保存到sp寄存器中	
//根据下面的board_init_f_alloc_reserve函数
//函数返回值为分配gb后的指针位置
```
board_init_f_alloc_reserve函数原型如下：
```
ulong board_init_f_alloc_reserve(ulong top)
{
	//将栈顶指针传进来，栈顶指针减去全局变量的长度意味着数据入栈，即在栈里预留变量存储空间。
	/* Reserve early malloc arena */
#if defined(CONFIG_SYS_MALLOC_F)
	top -= CONFIG_SYS_MALLOC_F_LEN;
#endif
	/* LAST : reserve GD (rounded up to a multiple of 16 bytes) */
	top = rounddown(top-sizeof(struct global_data), 16);
	return top;
}
```
## 3. 在栈中gd空间清零
```
	mov	r9, r0	//将栈顶指针存到r9寄存器里面,方便后续设置gd指针
	bl	board_init_f_init_reserve	//全局数据全部清零
```
board_init_f_init_reserve	函数定义如下：
```
void board_init_f_init_reserve(ulong base)
{
	struct global_data *gd_ptr;
#ifndef _USE_MEMCPY
	int *ptr;
#endif

	/*
	 * clear GD entirely and set it up.
	 * Use gd_ptr, as gd may not be properly set yet.
	 * 清除GD分配空间
	 */
	
	gd_ptr = (struct global_data *)base;
	/* zero the area */
#ifdef _USE_MEMCPY
	memset(gd_ptr, '\0', sizeof(*gd));	//全局数据区全部清零
#else
	for (ptr = (int *)gd_ptr; ptr < (int *)(gd_ptr + 1); )
		*ptr++ = 0;
#endif

	/* set GD unless architecture did it already */
#if !defined(CONFIG_ARM)
	arch_setup_gd(gd_ptr);
#endif
	/* next alloc will be higher by one GD plus 16-byte alignment */
	base += roundup(sizeof(struct global_data), 16);

	/*
	 * record early malloc arena start.
	 * Use gd as it is now properly set for all architectures.
	 */

#if defined(CONFIG_SYS_MALLOC_F)
	/* go down one 'early malloc arena' */
	gd->malloc_base = base;
	/* next alloc will be higher by one 'early malloc arena' size */
	base += CONFIG_SYS_MALLOC_F_LEN;
#endif
}
```

##4. 调用board_init_f，初始化各种硬件

```
	mov	r0, #0
	bl	board_init_f	// jump to ==> board_f.c
```
在board_init_f函数中所进行的主要操作如下，其中init_sequence_f[ ]是一个数组，其内容为一系列初始化函数，在函数initcall_run_list中依次调用init_sequence_f数组的各个初始化函数。

```
void board_init_f(ulong boot_flags)
{
	//此处省略多行代码
	gd->flags = boot_flags;
	gd->have_console = 0;
	//通过调用initcall_run_list函数，执行各项初始化	
	if (initcall_run_list(init_sequence_f)) 
		hang();
}
```
initcall_run_list函数如下：
```
int initcall_run_list(const init_fnc_t init_sequence[])
{	
	const init_fnc_t *init_fnc_ptr;
	for (init_fnc_ptr = init_sequence; *init_fnc_ptr; ++init_fnc_ptr) {
		int ret;
		//此处省略很多代码
		//通过函数指针依次调用数组内函数
		ret = (*init_fnc_ptr)();		
	}
return 0;
}
```
数组init_sequence_f[ ]定义如下，每个成员为一个函数指针，函数参数为void，返回类型为int
```
static init_fnc_t init_sequence_f[] = {
#ifdef CONFIG_SANDBOX
	setup_ram_buf,
#endif
	setup_mon_len,
#ifdef CONFIG_OF_CONTROL
	fdtdec_setup,
#endif
#ifdef CONFIG_TRACE
	trace_early_init,
#endif
	initf_malloc,
	initf_console_record,
#if defined(CONFIG_MPC85xx) || defined(CONFIG_MPC86xx)
	/* TODO: can this go into arch_cpu_init()? */
	probecpu,
#endif
#if defined(CONFIG_X86) && defined(CONFIG_HAVE_FSP)
	x86_fsp_init,
#endif
	arch_cpu_init,		/* basic arch cpu dependent setup */
	mach_cpu_init,		/* SoC/machine dependent CPU setup */
	initf_dm,
	arch_cpu_init_dm,
	mark_bootstage,		/* need timer, go after init dm */
#if defined(CONFIG_BOARD_EARLY_INIT_F)
	board_early_init_f,
#endif
	/* TODO: can any of this go into arch_cpu_init()? */
#if defined(CONFIG_PPC) && !defined(CONFIG_8xx_CPUCLK_DEFAULT)
	get_clocks,		/* get CPU and bus clocks (etc.) */
#if defined(CONFIG_TQM8xxL) && !defined(CONFIG_TQM866M) \
		&& !defined(CONFIG_TQM885D)
	adjust_sdram_tbs_8xx,
#endif
	/* TODO: can we rename this to timer_init()? */
	init_timebase,
#endif
#if defined(CONFIG_ARM) || defined(CONFIG_MIPS) || \
		defined(CONFIG_BLACKFIN) || defined(CONFIG_NDS32) || \
		defined(CONFIG_SH) || defined(CONFIG_SPARC)
	timer_init,		/* initialize timer */
#endif
#ifdef CONFIG_SYS_ALLOC_DPRAM
#if !defined(CONFIG_CPM2)
	dpram_init,
#endif
#endif
#if defined(CONFIG_BOARD_POSTCLK_INIT)
	board_postclk_init,
#endif
#if defined(CONFIG_SYS_FSL_CLK) || defined(CONFIG_M68K)
	get_clocks,
#endif
	env_init,		/* initialize environment */
#if defined(CONFIG_8xx_CPUCLK_DEFAULT)
	/* get CPU and bus clocks according to the environment variable */
	get_clocks_866,
	/* adjust sdram refresh rate according to the new clock */
	sdram_adjust_866,
	init_timebase,
#endif
	init_baud_rate,		/* initialze baudrate settings */
	serial_init,		/* serial communications setup */
	console_init_f,		/* stage 1 init of console */
#ifdef CONFIG_SANDBOX
	sandbox_early_getopt_check,
#endif
	display_options,	/* say that we are here */
	display_text_info,	/* show debugging info if required */
#if defined(CONFIG_MPC8260)
	prt_8260_rsr,
	prt_8260_clks,
#endif /* CONFIG_MPC8260 */
#if defined(CONFIG_MPC83xx)
	prt_83xx_rsr,
#endif
#if defined(CONFIG_PPC) || defined(CONFIG_M68K) || defined(CONFIG_SH)
	checkcpu,
#endif
	print_cpuinfo,		/* display cpu info (and speed) */
#if defined(CONFIG_MPC5xxx)
	prt_mpc5xxx_clks,
#endif /* CONFIG_MPC5xxx */
#if defined(CONFIG_DTB_RESELECT)
	embedded_dtb_select,
#endif
#if defined(CONFIG_DISPLAY_BOARDINFO)
	show_board_info,
#endif
	INIT_FUNC_WATCHDOG_INIT
#if defined(CONFIG_MISC_INIT_F)
	misc_init_f,
#endif
	INIT_FUNC_WATCHDOG_RESET
#if defined(CONFIG_HARD_I2C) || defined(CONFIG_SYS_I2C)
	init_func_i2c,
#endif
#if defined(CONFIG_HARD_SPI)
	init_func_spi,
#endif
	announce_dram_init,
	/* TODO: unify all these dram functions? */
#if defined(CONFIG_ARM) || defined(CONFIG_X86) || defined(CONFIG_NDS32) || \
		defined(CONFIG_MICROBLAZE) || defined(CONFIG_AVR32) || \
		defined(CONFIG_SH)
	dram_init,		/* configure available RAM banks */
#endif
#if defined(CONFIG_MIPS) || defined(CONFIG_PPC) || defined(CONFIG_M68K)
	init_func_ram,
#endif
#ifdef CONFIG_POST
	post_init_f,
#endif
	INIT_FUNC_WATCHDOG_RESET
#if defined(CONFIG_SYS_DRAM_TEST)
	testdram,
#endif /* CONFIG_SYS_DRAM_TEST */
	INIT_FUNC_WATCHDOG_RESET

#ifdef CONFIG_POST
	init_post,
#endif
	INIT_FUNC_WATCHDOG_RESET
	/*
	 * Now that we have DRAM mapped and working, we can
	 * relocate the code and continue running from DRAM.
	 *
	 * Reserve memory at end of RAM for (top down in that order):
	 *  - area that won't get touched by U-Boot and Linux (optional)
	 *  - kernel log buffer
	 *  - protected RAM
	 *  - LCD framebuffer
	 *  - monitor code
	 *  - board info struct
	 */
	setup_dest_addr,
#if defined(CONFIG_BLACKFIN) || defined(CONFIG_XTENSA)
	/* Blackfin u-boot monitor should be on top of the ram */
	reserve_uboot,
#endif
#if defined(CONFIG_SPARC)
	reserve_prom,
#endif
#if defined(CONFIG_LOGBUFFER) && !defined(CONFIG_ALT_LB_ADDR)
	reserve_logbuffer,
#endif
#ifdef CONFIG_PRAM
	reserve_pram,
#endif
	reserve_round_4k,
#if !(defined(CONFIG_SYS_ICACHE_OFF) && defined(CONFIG_SYS_DCACHE_OFF)) && \
		defined(CONFIG_ARM)
	reserve_mmu,
#endif
#ifdef CONFIG_DM_VIDEO
	reserve_video,
#else
# ifdef CONFIG_LCD
	reserve_lcd,
# endif
	/* TODO: Why the dependency on CONFIG_8xx? */
# if defined(CONFIG_VIDEO) && (!defined(CONFIG_PPC) || defined(CONFIG_8xx)) && \
		!defined(CONFIG_ARM) && !defined(CONFIG_X86) && \
		!defined(CONFIG_BLACKFIN) && !defined(CONFIG_M68K)
	reserve_legacy_video,
# endif
#endif /* CONFIG_DM_VIDEO */
	reserve_trace,
#if !defined(CONFIG_BLACKFIN) && !defined(CONFIG_XTENSA)
	reserve_uboot,
#endif
#ifndef CONFIG_SPL_BUILD
	reserve_malloc,
	reserve_board,
#endif
	setup_machine,
	reserve_global_data,
	reserve_fdt,
	reserve_arch,
	reserve_stacks,
	setup_dram_config,
	show_dram_config,
#if defined(CONFIG_M68K) || defined(CONFIG_MIPS) || defined(CONFIG_PPC) || \
	defined(CONFIG_SH)
	setup_board_part1,
#endif
#if defined(CONFIG_PPC) || defined(CONFIG_M68K)
	INIT_FUNC_WATCHDOG_RESET
	setup_board_part2,
#endif
	display_new_sp,
#ifdef CONFIG_SYS_EXTBDINFO
	setup_board_extra,
#endif
	INIT_FUNC_WATCHDOG_RESET
	reloc_fdt,
	setup_reloc,
#if defined(CONFIG_X86) || defined(CONFIG_ARC)
	copy_uboot_to_ram,
	clear_bss,
	do_elf_reloc_fixups,
#endif
#if defined(CONFIG_XTENSA)
	clear_bss,
#endif
#if !defined(CONFIG_ARM) && !defined(CONFIG_SANDBOX)
	jump_to_copy,
#endif
	NULL,
};
```




