#define MAX_PAYLOAD_LEN 250

#define FIND_DEVICES		72
#define NEED_ADDR			73
#define ACK_DEVICE			74
#define ACK_REPLY			75

#define INTERROGATE			80
#define INTERROGATE_REPLY 	81

#define PING 				44
#define PING_REPLY 			45

#define RESET_NETWORK 		58

#define GET 				36
#define GET_REPLY 			37
#define SET 				38
#define SET_REPLY 			39

#define UNKNOWN_REPLY		33

struct mheader_t {
	uint8_t destination;
	uint8_t mtype;
	uint16_t length;
	uint32_t mac;
	uint16_t crc;
} __attribute__ ((packed));

struct message_t {
	struct mheader_t header;
	char * payload;
	uint16_t payload_len;
} __attribute__ ((packed));

void gb_push_char(char *buffer, char in);