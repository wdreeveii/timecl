#include <inttypes.h>
#include <stddef.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#include <avr/io.h>
#include <util/atomic.h>

#include "config.h"
#include "usart.h"
#include "iocontrol.h"


#define SERIALNUM_LENGTH		8

struct s_config {
	uint16_t			S1Baud;
	uint16_t			S2Baud;
	int8_t				SerialNumber[SERIALNUM_LENGTH];
	uint8_t				portconfig[NUM_PORTS];
};

uint8_t EEPROM_read(uint16_t uiAddress)
{
	/* Wait for completion of previous write */
	while(EECR & (1<<EEPE)) ;
	/* Set up address register */
	EEAR = uiAddress;
	/* Start eeprom read by writing EERE */
	EECR |= (1<<EERE);
	/* Return data from Data Register */
	return EEDR;
}

void EEPROM_read_page(uint16_t start_address, uint8_t *data, uint16_t length)
{
	uint16_t index = 0;
	for (; index < length; ++index)
	{
		*(data + index) = EEPROM_read(start_address + index);
	}
}

void EEPROM_write(uint16_t uiAddress, uint8_t ucData)
{
	/* Wait for completion of previous write */
	while(EECR & (1<<EEPE));
	
	/* Set up address and Data Registers */
	EEAR = uiAddress;
	EEDR = ucData;
	
	/* Write logical one to EEMPE */
	EECR |= (1<<EEMPE);
	/* Start eeprom write by setting EEPE */
	EECR |= (1<<EEPE);
}

void EEPROM_write_page(uint16_t start_address, uint8_t *data, uint16_t length)
{
	uint16_t index = 0;
	for(;index < length; ++index)
	{
		EEPROM_write(start_address + index,  *(data + index));	
	}	
}

void config_get_io_types(uint8_t *data)
{
	EEPROM_read_page(offsetof(struct s_config, portconfig), data, sizeof(uint8_t) * NUM_PORTS);
}
void config_set_io_types(uint8_t *data)
{
	EEPROM_write_page(offsetof(struct s_config, portconfig), data, sizeof(uint8_t) * NUM_PORTS);
}

void config_Init()
{

}


uint16 config_get_baud(uint8 port)
{
	uint16 baud = 0;
	switch(port)
	{
		case 0:
			baud = EEPROM_read(offsetof(struct s_config, S1Baud)) << 8;
			baud |= EEPROM_read(offsetof(struct s_config, S1Baud) + 1);
			break;
		case 1:
			baud = EEPROM_read(offsetof(struct s_config, S2Baud)) << 8;
			baud |= EEPROM_read(offsetof(struct s_config, S2Baud) + 1);
			break;
	}	
	return baud;
}

void config_set_baud(uint8 port, uint16 baud)
{
	switch (port)
	{
		case 0:
			EEPROM_write(offsetof(struct s_config, S1Baud), baud << 8);
			EEPROM_write(offsetof(struct s_config, S1Baud) + 1, baud);
			break;
		case 1:
			EEPROM_write(offsetof(struct s_config, S2Baud), baud << 8);
			EEPROM_write(offsetof(struct s_config, S2Baud) + 1, baud);
			break;	
	}	
}
