package jobs

import (
	"os"
	"time"
	"fmt"
	"bytes"
	"encoding/binary"
	//"github.com/robfig/revel"
	//"github.com/mewkiz/pkg/hashutil/crc16"
	//"github.com/robfig/revel/modules/jobs/app/jobs"
	"github.com/wdreeveii/termioslib"
)

type header_t struct {
    destination uint8
    mtype uint8
    length uint16
    mac uint32
    crc uint16
}

type message_t struct {
    header header_t
    payload []byte
}

func crc16_update(crc uint16, a byte) uint16 {
    var i int
    crc ^= uint16(a)
    for i = 0; i < 8; i++ {
	if (crc & 1 != 0) {
	    crc = (crc >> 1) ^ 0xA001
	} else {
	    crc = (crc >> 1)
	}
    }
    return crc
}

func ToByte(msg message_t) []byte {
    b := new(bytes.Buffer)
    b.Write([]byte("A"))
    binary.Write(b, binary.LittleEndian, msg.header)
    b.Write([]byte("A"))
    b.Write(msg.payload)
    bytes := b.Bytes()
    var checksum uint16 = 0xffff
    for i := 0; i < len(bytes); i++ {
	checksum = crc16_update(checksum, bytes[i])
    }
    binary.Write(b, binary.LittleEndian, checksum)
    return b.Bytes()
}

func HeaderCRC(msg *message_t) uint16 {
    msg.header.length = uint16(len(msg.payload) + 2) // add space for the crc
    b := new(bytes.Buffer)
    b.Write([]byte("A"))
    binary.Write(b, binary.LittleEndian, msg.header)
    bytes := b.Bytes()
    var checksum uint16 = 0xffff
    for i := 0; i < len(bytes) - 2; i++ {
	checksum = crc16_update(checksum, bytes[i])
    }
    return checksum
}

func runmaster() {
    var (
	err error
	orig_termios termioslib.Termios
	work_termios termioslib.Termios
	ser *os.File
    )
    
    defer func () {
	fmt.Println("Serial Exiting...")
	if err != nil { 
	    fmt.Println(err)
	}
    }()
    
    ser, err = os.OpenFile("/dev/ttyUSB0", os.O_RDWR, 777)
    if err != nil { return }
    
    defer func () {
	fmt.Println("Closing serial port")
	ser.Close()
    }()
    
    
    
    // flush all buffers
    if err = termioslib.Flush(ser.Fd(), termioslib.TCIOFLUSH); err != nil { return }
    
    // save a copy of the original terminal configuration
    if err = termioslib.Getattr(ser.Fd(), &orig_termios); err != nil { return }

    // get a working copy
    if err = termioslib.Getattr(ser.Fd(), &work_termios); err != nil { return }

    work_termios.C_iflag &= ^(termioslib.IGNBRK | termioslib.BRKINT | termioslib.IGNPAR | termioslib.PARMRK | termioslib.INPCK | termioslib.ISTRIP | termioslib.INLCR | termioslib.IGNCR | termioslib.ICRNL | termioslib.IXON | termioslib.IXOFF | termioslib.IXANY | termioslib.IMAXBEL | termioslib.IUTF8)
    work_termios.C_oflag &= ^(termioslib.OPOST | termioslib.ONLCR )
    work_termios.C_lflag &= ^(termioslib.ISIG | termioslib.ICANON | termioslib.IEXTEN | termioslib.ECHO | termioslib.ECHOE | termioslib.ECHOK | termioslib.ECHONL | termioslib.NOFLSH | termioslib.TOSTOP | termioslib.ECHOPRT | termioslib.ECHOCTL | termioslib.ECHOKE)
    work_termios.C_cflag &= ^(termioslib.CSIZE | termioslib.PARENB | termioslib.PARODD | termioslib.HUPCL | termioslib.CSTOPB | termioslib.CRTSCTS)
    work_termios.C_cflag |= (termioslib.CS8 | termioslib.CREAD | termioslib.CLOCAL)
    work_termios.C_cc[termioslib.VMIN] = 1
    work_termios.C_cc[termioslib.VTIME] = 0

    termioslib.Setspeed(&work_termios, termioslib.B57600)
    // set the working copy
    if err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW , &work_termios); err != nil { return }

    // set the settings back to the original when the program exits
    defer func () {
	fmt.Println("Resetting termios")
        err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW, &orig_termios)
    } ()

    w:= make(chan message_t)
    go WriteMessage(ser, w)
    for {
	defer func () {
	    fmt.Println("Master Loop Exiting...")
	    if err != nil { 
		fmt.Println(err)
	    }
	}()
	var msg message_t
	msg.payload = []byte("HAHAHA")
	
	msg.header.destination = 1
	msg.header.mtype = 45
	msg.header.mac = 6
	
	w <- msg
	fmt.Println("beat")
	time.Sleep(1*time.Second)
    }
}

func WriteMessage(c *os.File, w chan message_t) {
    var msg message_t
    for {
	msg = <- w
	msg.header.crc = HeaderCRC(&msg)
	fmt.Printf("msglen: %d\n", msg.header.length)
	fmt.Printf("Msg To: %d Type: %d\n", msg.header.destination, msg.header.mtype)
	_, err := c.Write(ToByte(msg))
	if err != nil {
	    fmt.Println(err)
	    //return
	}
    }
}
    
func init() {
	fmt.Println("blah")
	go runmaster() 
}
