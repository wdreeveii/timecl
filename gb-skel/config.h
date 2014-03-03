
#ifndef _CONFIG_H
#define _CONFIG_H

#define F_CPU 20000000UL
// Inline assembly: The nop = do nothing for one clock cycle.
#define nop()  __asm__ __volatile__("nop")
#include <inttypes.h>

#define PACKED __attribute__ ((packed))
//typedef _Bool     bool;
typedef _Bool     bool_t;

typedef uintptr_t ptr_t;

// Define some useful types:
typedef signed char int8;
typedef uint8_t uint8;
typedef uint16_t uint16;
typedef uint32_t uint32;

#define DBG8 "%02x"
#define DBG16 "%04x"
#define DBG32 "%08lx"

#define DEBUGF printf

#define TRUE 1
#define FALSE 0

#define MUCRON_EVENTLIST_SIZE		16

#define NELEM(x) (sizeof(x)/sizeof(x[0]))

void config_Init();
void config_get_io_types(uint8_t *data);
void config_set_io_types(uint8_t *data);

uint16 config_get_baud(uint8 port);
void config_set_baud(uint8 port, uint16 baud);

#endif
