#include "config.h"
#include <avr/io.h>
#include <avr/interrupt.h>
#include <util/atomic.h>
#include <util/delay.h>
#include "buffer.h"
#include "usart.h"
#include "iocontrol.h"
#include "gb_client.h"
#include <string.h>
#include <limits.h>
#include <stdlib.h>
#include <stdio.h>


static int putchar_for_printf(char c, FILE *stream)
{
	/* print also \r if the character is \n */
	if (c == '\n') putchar_for_printf('\r', stream);
		
	USART_Send(1, (uint8_t *) &c, 1);
	return 0;	
}

static FILE standard_output = FDEV_SETUP_STREAM(putchar_for_printf, NULL, _FDEV_SETUP_WRITE);

/*
 * Input and output buffers declarations
 */
buffer_t in_buf, out_buf[2];

/*
 *  USART_Init : Initialize USARTS
 *
 *  Purpose : Init the USART with chosen baudrate, number of
 *            data bits and stop bit and enable the transmitter
 *            & reciever, and enable the recieve interrupt.
 *            
 *            Also enable the buffers used to store data being
 *            transmitted and recieved by this usart.
 *
 *  Input : baudrate
 */
void USART_Init0(uint16 baud)
{
	/* set tx driver enable pin to output */
	DDRC |= 1;
	/* ouput low tx driver enable */
	PORTC &= ~1;

	/* Wait latest receive character was pull from UDR */
	while(UCSR0A & (1<<RXC0));

	/* Wait end of transmission */
	while(!(UCSR0A & (1<<UDRE0)));

	/* Disable transmitter and receiver */
	UCSR0B &= ~((1<<RXEN0)|(1<<TXEN0));

	/* Init buffers */
	Buffer_Reset(&in_buf);
	Buffer_Reset(&out_buf[0]);
	
	/* Set double transmission speed */
	UCSR0A = (1 << U2X0);
	
	/* Set baud rate */
	UBRR0H = (unsigned char)(baud>>8);
	UBRR0L = (unsigned char)baud;

	/* Set 8 bit frame size */
	UCSR0C = (3<<UCSZ00);
	
	/* Enable transmitter and receiver (RXEN and TXEN), enable reception interrupt (RXCIE) */
	UCSR0B = (1<<RXCIE0)|(1<<RXEN0)|(1<<TXEN0);
	
}

void USART_Init1(uint16 baud)
{
	/* Wait latest receive character was pull from UDR */
	while(UCSR1A & (1<<RXC1));

	/* Wait end of transmission */
	while(!(UCSR1A & (1<<UDRE1)));

	/* Disable transmitter and receiver */
	UCSR1B &= ~((1<<RXEN1)|(1<<TXEN1));
	
	Buffer_Reset(&out_buf[1]);
	
	/* Set double transmission speed */
	UCSR1A = (1 << U2X1);

	/* Set baud rate */
	UBRR1H = (unsigned char)(baud>>8);
	UBRR1L = (unsigned char)baud;
	
	/* Set 8 bit frame size */
	UCSR1C = (3<<UCSZ10);
	
	/* Enable transmitter and receiver (RXEN and TXEN), enable reception interrupt (RXCIE) */
	UCSR1B = (1<<RXCIE1)|(1<<RXEN1)|(1<<TXEN1);
}

/*
 *  USART_Init : Initialize USART
 *
 *  Purpose : Init the USART with baudrate stored in config eeprom,
 *            If no baudrate is found, use a default value of 42 aka 57600baud.
 */
void USART_Init(void)
{
	// default to 42 which is 57600 baud at 20mhz
	uint16 baud = 42;
	uint16 config_baud = 0;

	// The following lines are helpful when creating a config eeprom from scratch
	config_set_baud(0, 42);
	config_set_baud(1, 42);
	if ((config_baud = config_get_baud(0))) baud = config_baud;
	USART_Init0(baud);
	
	// setup stdout so printf works
	stdout = &standard_output;
	
	baud = 42;
	if ((config_baud = config_get_baud(1))) baud = config_baud;
	USART_Init1(baud);
}

/*
 *  Usart_Send : Send a frame
 *
 *  Purpose : Copy data to output buffer and enable
 *            UDR empty interrupt
 *
 *  Input : p_data, pointer on data frame to send
 *          length, length of frame
 */
void USART_Send(uint8_t port, uint8_t *p_data, uint16_t length)
{
	uint8_t i = 0;
	/* busy wait if we can not push onto the buffer */
	while (length)
	{
		for (i = 0; i < length; i++)
		{
			ATOMIC_BLOCK(ATOMIC_RESTORESTATE)
			{
				if (Buffer_Push(&(out_buf[!!port]), *(p_data + i)) == -1 )
				{	
					i -= 1;
					break;
				}
			}
		}
		p_data += i;
		length -= i;

		/* Enable UDR Empty interrupt */
		if (port == 0)
			UCSR0B |= (1<<UDRIE0);
		else
			UCSR1B |= (1<<UDRIE1);
	}
}

/*
 *  Interrupt Routine : UDR Empty
 *
 *  Purpose : Try to take a character from output buffer. If
 *            buffer is empty (character == -1), disable this
 *            interrupt. Else, put char in UDR.
 */
ISR(USART0_UDRE_vect)
{
	char c;	
	
	/* pull a byte out of the buffer */
	if(Buffer_Pull(&(out_buf[0]), (unsigned char *)&c) == -1) {
		/* If buffer empty stop UDR empty interrupt : Tx End */
		UCSR0B &= ~(1<<UDRIE0);
		/* disable driver tx */
		_delay_us(150);
		PORTC &= ~1;
	} else {
		/* Enable driver output */
		PORTC |= 1;
		/* Else, put c in UDR */
		UDR0 = c;
	}
}

ISR(USART1_UDRE_vect)
{
	char c;
	
	/* pull a byte out of the buffer */
	if(Buffer_Pull(&(out_buf[1]), (unsigned char *)&c) == -1)
		/* If buffer empty stop UDR empty interrupt : Tx End */
		UCSR1B &= ~(1<<UDRIE1);
	else
	{
		/* Else, put c in UDR */
		UDR1 = c;
		//printf("%x|", (unsigned char)c);	
	}
		
}

/*
 * Interrupt Routine : Receive complete
 *
 * Purpose : Check frame error and parity error. If there is at
 *           least one error, read character from UDR and put
 *           it in garbage. Else, push it into input buffer.
 *           Figure out how many bytes are in the packet and
 *           process the packet if all bytes are received.
 */
ISR(USART0_RX_vect)
{
	static char data[500];
	char garbage;
	if((UCSR0A & (1<<FE0))||(UCSR0A & (1<<UPE0)))
		/* If frame error or parity error UDR is Garbage */
		garbage = UDR0;
	else
	{
		/* Else, send received char into input buffer */
		garbage = UDR0;
		gb_push_char(data, garbage);
		//if (usart0_rx_handler != NULL) (*usart0_rx_handler)(0, &in_buf[0]);
	}
	
}

ISR(USART1_RX_vect)
{
	uint8_t garbage;
	if((UCSR1A & (1<<FE1))||(UCSR1A & (1<<UPE1)))
		/* If frame error or parity error UDR is Garbage */
		garbage = UDR1;
	else
	{
		/* Else, send received char into input buffer */
		garbage = UDR1;
		Buffer_Push(&in_buf, garbage);
	}
}

