#include "config.h"

#include <stdio.h>
#include <avr/io.h>
#include <avr/pgmspace.h>
#include <avr/interrupt.h>
#include <util/atomic.h>
#include <util/delay.h>
#include <string.h>
#include "iocontrol.h"
#include "usart.h"

void build_analog_tables();

struct io_port {
	uint8_t * regaddr;
	uint8_t * valaddr;
	uint8_t bit;
};

const struct io_port ports[NUM_PORTS] = {
	{(uint8_t *)&DDRB, (uint8_t *)&PINB, 0},
	{(uint8_t *)&DDRB, (uint8_t *)&PINB, 1},
	{(uint8_t *)&DDRB, (uint8_t *)&PINB, 2},
	{(uint8_t *)&DDRB, (uint8_t *)&PINB, 3},
	{(uint8_t *)&DDRB, (uint8_t *)&PINB, 4},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 1},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 7},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 6},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 5},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 4},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 3},
	{(uint8_t *)&DDRC, (uint8_t *)&PINC, 2},
	{(uint8_t *)&DDRD, (uint8_t *)&PIND, 4},
	{(uint8_t *)&DDRD, (uint8_t *)&PIND, 4},
	{(uint8_t *)&DDRD, (uint8_t *)&PIND, 4},
	{(uint8_t *)&DDRD, (uint8_t *)&PIND, 4},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 0},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 1},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 2},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 3},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 4},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 5},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 6},
	{(uint8_t *)&DDRA, (uint8_t *)&PINA, 7},

};

uint8_t iotypes[NUM_PORTS];
uint8_t adc_channels[8];
uint16_t adc_vals[8];
static uint8_t num_channels = 0;
static uint8_t current_channel = 0;

void io_init()
{		
	//config_get_io_types(iotypes);

	io_set_type(12, PORT_AOUTPUT);
	io_set_type(13, PORT_AOUTPUT);
	io_set_type(14, PORT_AOUTPUT);
	io_set_type(15, PORT_AOUTPUT);

	io_set_type(16, PORT_AINPUT);
	io_set_type(17, PORT_AINPUT);
	io_set_type(18, PORT_AINPUT);
	io_set_type(19, PORT_AINPUT);
	io_set_type(20, PORT_AINPUT);
	io_set_type(21, PORT_AINPUT);
	io_set_type(22, PORT_AINPUT);
	io_set_type(23, PORT_AINPUT);
	
	build_analog_tables();

	if (num_channels)
	{
		ADCSRA = (1 << ADPS2) | (1 << ADPS1) | (1 << ADPS0);
		ADMUX = (1 << REFS0) | (1 << adc_channels[0]);
		ADCSRA |= (1 << ADATE);
		ADCSRA |= (1 << ADEN);
		ADCSRA |= (1 << ADIE);
		ADCSRA |= (1 << ADSC);
		
	}
}

void build_analog_tables()
{
	uint8_t index = 0;
	num_channels = 0;
	current_channel = 0;
	for(; index < NUM_PORTS; index++)
	{
		// Only AnalogINPUTS are on port A
		if (iotypes[index] == PORT_AINPUT && ports[index].regaddr == &DDRA)
		{
			DIDR0 |= (1 << ports[index].bit);
			adc_channels[num_channels++] = ports[index].bit;
		}
	}
}
void io_set_type(uint8_t port, uint8_t type)
{	
	if (port >= NUM_PORTS)
		return;

	if (type == PORT_AINPUT)
	{
		// make sure on port a
		if (ports[port].regaddr == &DDRA)
		{
			ATOMIC_BLOCK(ATOMIC_RESTORESTATE)
			{
				iotypes[port] = type;
				build_analog_tables();
			}
		}
	}
	else
	{
		ATOMIC_BLOCK(ATOMIC_RESTORESTATE)
		{
			iotypes[port] = type;
			if (ports[port].regaddr == &DDRA)
				build_analog_tables();
		}
	}
}

uint8_t io_num_ports() {
	return NUM_PORTS;
}

uint8_t io_get_type(uint8_t port) {
	uint8_t type;
	if (port >= NUM_PORTS)
		return 0;

	ATOMIC_BLOCK(ATOMIC_RESTORESTATE) {
		type = iotypes[port];
	}
	return type;
}

ISR(ADC_vect, ISR_BLOCK)
{
	static uint8_t dirty = 2;
	uint8_t value = 0;
	if (!dirty && num_channels)
	{
		// save the value
		value = ADCL;
		value |= (ADCH << 8);
		adc_vals[current_channel++] = value << 6;
		
		if (current_channel == num_channels)
			current_channel = 0;
		// set mux to new current_channel
		ADMUX = (1 << REFS0) | adc_channels[current_channel];
		// do some shit
		dirty = 2;
	}
	dirty--;
		
}
uint8_t io_read(uint8_t index)
{
	if (index >= NUM_PORTS)
		return 0;
		
	if (iotypes[index] == PORT_BINPUT)
	{
		return !!(*(ports[index].valaddr) & (1U << ports[index].bit));
	}
	return 0;
}

uint16_t io_aread(uint8_t index)
{
	uint16_t copy;
	if (index >= NUM_PORTS)
		return 0;
	
	if (iotypes[index] == PORT_AINPUT)
	{
		index = ports[index].bit;
		
		ATOMIC_BLOCK(ATOMIC_RESTORESTATE)
		{
			copy = adc_vals[index];
		}
		return copy;
	}
	return 0;
}

void iocontrol(uint8_t port, uint8_t on)
{
	if (port >= NUM_PORTS || iotypes[port] != PORT_OUTPUT)
		return;
		
	struct io_port ctrl;
	ctrl.regaddr = ports[port].regaddr;
	ctrl.bit = ports[port].bit;
	
	if (on)
		*(ctrl.regaddr) |= (1U<<(ctrl.bit));
	else
		*(ctrl.regaddr) &= ~(1U<<(ctrl.bit));
}

void ioflip(uint8_t port)
{
	if (port >= NUM_PORTS || iotypes[port] != PORT_OUTPUT)
		return;
	struct io_port ctrl;
	ctrl.regaddr = ports[port].regaddr;
	ctrl.bit = ports[port].bit;
	
	unsigned char state = (*(ctrl.regaddr)) & (1U<<(ctrl.bit));
	
	if (state)
		*(ctrl.regaddr) &= ~(1U<<(ctrl.bit));
	else
		*(ctrl.regaddr) |= (1U<<(ctrl.bit));
}
