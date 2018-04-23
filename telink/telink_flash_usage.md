## telink flash usage

### flash 地址定义

```
    flash_adr_mac 			= 0x76000;			\
    flash_adr_pairing 		= 0x77000;		\
    flash_adr_dev_grp_adr   = 0x79000;    \
    flash_adr_lum 			= 0x78000;			\
    flash_adr_ota_master 	= 0x20000;     \
    flash_adr_reset_cnt 	= 0x7A000;      \
    flash_adr_alarm 	    = 0x7B000;          \
    flash_adr_scene 	    = 0x7C000;          \
    flash_adr_user_data   = 0x7D000;          \
    flash_adr_light_new_fw  = 0x40000;
```

512k的flash有8个64-K的block，每个block有16个4K的sector，如下图所示。可以看出mac、pariring、group_addr、lum、reset_cnt、alarm、scene、user_data等数据是放在flash最后一个block的位置；ota_master在block[2]的位置，新的固件在block[4]的位置。

![flash_arch](.\pic\flash_arch.png)




