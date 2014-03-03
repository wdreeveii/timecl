// Firmware for fdio controller V1
#include "config.h"

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <avr/io.h>
#include <avr/interrupt.h>
#include <util/delay.h>
#include <util/atomic.h>
#include <math.h>

#include "usart.h"
#include "util.h"
#include "iocontrol.h"

static void hardware_init()
{
	DDRA = 0;
	DDRB = 0;
	DDRC = 0xC0;
	DDRD = 0;
	PORTA = 0;
	PORTB = 0;
	PORTC = 0xC0;
	PORTD = 0;
	io_init();
	config_Init();
	USART_Init();
}

int main(void)
{
	uint8_t mcusr;
	cli();
	hardware_init();
	mcusr = MCUSR;
	MCUSR = 0;
	sei();
	DSEND("Precision Horticulture\n");
	DSEND("Model: GB Proving Grounds 1.0\n");
	printf("mcusr %x\n", mcusr);
	
	
	while (1)
	{
	}	
	// Never reached.
	return(0);
}

ISR(BADISR_vect)
{
   DDRC |= (1U<<5);
}

