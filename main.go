package main

import "errors"

func main() {}

/*
	 Char. number range  |        UTF-8 octet sequence
			(hexadecimal)    |              (binary)
	 --------------------+---------------------------------------------
	 0000 0000-0000 007F | 0xxxxxxx
	 0000 0080-0000 07FF | 110xxxxx 10xxxxxx
	 0000 0800-0000 FFFF | 1110xxxx 10xxxxxx 10xxxxxx
	 0001 0000-0010 FFFF | 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
*/

// -- BITWISE AND (&)
//	0011
//	0101
//	0001

//	110xxxxx
//	11100000
//	11000000

//	1110xxxx
//	11110000
//	11100000
// -----

// 110xxxxx
// 00011111

// Uma runa ocupa 4 byes (32 bits) na memória. Ou seja, segue este padrão:
// xxxxxxxx xxxxxxxx xxxxxxxx xxxxxxxx
// Ao transformar os bytes em runa, o Golang (e grande parte das linguagens fortemente tipadas)
// Ignorará os 3 primeiros bytes e considerará apenas o ultimo. Neste padrão:
// 00000000 00000000 00000000 xxxxxxxx
// Desta forma, quanto a runa é de apenas 1 byte no padrão UTF-8, ela está dentro da tabela ASCII e
// pode ser convertida para runa diretamente.

//	Porém, em runas maiores que um byte, é necessário utilizar o BITWISE left-shift
// 	para "empurrar" os bites que definem o tamanho do byte
// 	Desta forma, os bites "andarão" a casa 6x exemplo:
// 00000000 00000000 00000000 000xxxxx
// 00000000 00000000 00000000 00xxxxx0
// 00000000 00000000 00000000 0xxxxx00
// 00000000 00000000 00000000 xxxxx000
// 00000000 00000000 0000000x xxxx0000
// 00000000 00000000 000000xx xxx00000
// 00000000 00000000 00000xxx xx000000
//
// 														10000000
// 														00111111
//
//	00000000 00000000 00000xxx 00xxxxxx

func decodeRune(b []byte) (r rune, s int, err error) {
	if len(b) == 0 {
		return 0, 0, errors.New("empty input")
	}
	b0 := b[0]

	switch {
	case b0 < 0x80: //	ASCII (1 byte caracter)
		if len(b) > 1 {
			if b[1]&0xC0 == 0x80 {
				return 0, 0, errors.New("invalid length")
			}
		}

		r = rune(b0)
		s = 1

	case b0&0xE0 == 0xC0: //2 byte caracter
		if len(b) < 2 {
			return 0, 0, errors.New("invalid length")
		}

		if len(b) > 2 {
			if b[2]&0xC0 == 0x80 {
				return 0, 0, errors.New("invalid length")
			}
		}

		b1 := b[1]

		if b1&0xC0 != 0x80 {
			return 0, 0, errors.New("invalid continuation byte")
		}

		s = 2
		r = ((rune(b0) & 0x1F) << 6) |
			(rune(b1 & 0x3F))

		if r < 0x80 {
			return 0, 0, errors.New("overlong")
		}

	case b0&0xF0 == 0xE0: //3 byte caracter
		if len(b) < 3 {
			return 0, 0, errors.New("invalid length")
		}

		if len(b) > 3 {
			if b[3]&0xC0 == 0x80 {
				return 0, 0, errors.New("invalid length")
			}
		}

		b1 := b[1]
		b2 := b[2]

		if b0 == 0xE0 && b1 < 0xA0 {
			return 0, 0, errors.New("overlong")
		}

		if b1&0xC0 != 0x80 || b2&0xC0 != 0x80 {
			return 0, 0, errors.New("invalid continuation byte")
		}

		r = ((rune(b0) & 0x0F) << 12) |
			((rune(b1 & 0x3F)) << 6) |
			(rune(b2) & 0x3F)
		s = 3

	case b0&0xF8 == 0xF0: //4 byte caracter
		if len(b) < 4 {
			return 0, 0, errors.New("invalid length")
		}

		if len(b) > 4 {
			if b[4]&0xC0 == 0x80 {
				return 0, 0, errors.New("invalid length")
			}
		}

		b1 := b[1]
		b2 := b[2]
		b3 := b[3]

		if b0 == 0xF0 && b1 < 0x90 {
			return 0, 0, errors.New("overlong")
		}

		if b1&0xC0 != 0x80 || b2&0xC0 != 0x80 || b3&0xC0 != 0x80 {
			return 0, 0, errors.New("invalid continuation byte")
		}

		r = ((rune(b0) & 0x07) << 18) |
			((rune(b1 & 0x3F)) << 12) |
			((rune(b2 & 0x3F)) << 6) |
			(rune(b3) & 0x3F)
		s = 4
	default:
		return 0, 0, errors.New("invalid utf-8")
	}

	if r >= 0xD800 && r <= 0xDFFF {
		return 0, 0, errors.New("surrogate halfs")
	}

	if r > 0x10FFFF {
		return 0, 0, errors.New("too big")
	}

	return r, s, nil
}
