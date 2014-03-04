AVRDUDE_PROGRAMMER	:= avrisp2
TARGET				:= fdioCV2.hex
ELF					:= fdioCV2.elf
SRCS				:= $(wildcard *.c)
CC					:= avr-gcc
OBJCOPY				:= avr-objcopy
OBJDUMP				:= avr-objdump

CCFLAGS = -std=c99 -mmcu=atmega1284p -O3 -Wall -fno-strict-aliasing
AVFLAGS = -c ${AVRDUDE_PROGRAMMER} -p m1284p -P usb
OCFLAGS = -j .text -j .data -j .boot -O ihex

.PHONY: all clean distclean 
all:: ${TARGET} 


${TARGET}: ${ELF}
	${OBJCOPY} ${OCFLAGS} $< $@
	
${ELF}: ${SRCS} 
	${CC} ${CCFLAGS} ${LDFLAGS} -o $@ ${SRCS} 

clean:: 
	-rm -f *~ *.o *.dep ${TARGET} ${ELF}
program: ${TARGET}
	avrdude ${AVFLAGS} -U flash:w:${TARGET} -U eeprom:w:config.hex
		
eeprom:
	avrdude ${AVFLAGS} -U eeprom:w:config.hex
	
fuse:
	avrdude ${AVFLAGS} -U lfuse:w:0xF7:m -U hfuse:w:0xD9:m -U efuse:w:0xFF:m

distclean:: clean
