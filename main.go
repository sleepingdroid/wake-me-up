// main.go
package main

import (
	"fmt"
	"time" // เพิ่ม import time สำหรับ Sleep

	"github.com/gopherjs/gopherjs/js"
)

func main() {
	// ใช้ channel เพื่อรัน Go โปรแกรมแบบไม่จบเมื่อถูกโหลดเป็น WASM
	// โดยปกติ main function ใน Go จะจบไปเลยถ้าไม่มีโค้ดที่รันค้าง
	c := make(chan struct{})
	fmt.Println("Go WebAssembly Stupid Group Alarm Initialized!")

	// กำหนดฟังก์ชัน Go "speakAlarm" ให้เป็น global object ใน JavaScript
	// ทำให้โค้ด JavaScript สามารถเรียกใช้ฟังก์ชันนี้ได้โดยตรง

	js.Global.Set("speakAlarm", js.MakeFunc(speakAlarmFunc))

	// บล็อก main goroutine ไม่ให้จบ เพื่อให้ Go-WASM ยังคงทำงานและพร้อมรับคำสั่งจาก JS
	<-c
}

// speakAlarmFunc เป็นฟังก์ชันที่จะถูกเรียกจาก JavaScript
// รับพารามิเตอร์ 2 ตัวจาก JS: ข้อความ (string) และความดัง (float64)
func speakAlarmFunc(this *js.Object, args []*js.Object) interface{} {
	// ตรวจสอบว่ามีข้อความถูกส่งมาหรือไม่
	if len(args) < 1 {
		fmt.Println("speakAlarm: No message provided.")
		return nil
	}
	message := args[0].String() // แปลง JS Value เป็น Go string
	volume := 1.0               // กำหนดค่าเริ่มต้นของความดัง

	// ตรวจสอบว่ามีความดังถูกส่งมาและเป็น Type Number
	if len(args) > 1 && args[1].Float() != 0 {
		volume = args[1].Float() // แปลง JS Value เป็น Go float64
		// จำกัดค่าความดังให้อยู่ในช่วง 0.0 ถึง 1.0
		if volume < 0.0 {
			volume = 0.0
		} else if volume > 1.0 {
			volume = 1.0
		}
	}

	fmt.Printf("Speaking: \"%s\" with volume %.2f\n", message, volume)

	// เข้าถึง Web Speech API (SpeechSynthesis) ของเบราว์เซอร์ผ่าน js.Global()
	speechSynthesis := js.Global.Get("speechSynthesis")
	if speechSynthesis == nil { // ตรวจสอบว่า API มีอยู่หรือไม่
		fmt.Println("Web Speech API not available.")
		return nil
	}

	// หยุดการพูดที่ค้างอยู่ก่อนหน้า เพื่อป้องกันเสียงซ้อนทับกัน
	speechSynthesis.Call("cancel")
	// หน่วงเวลาเล็กน้อยเพื่อให้เบราว์เซอร์ประมวลผลการ cancel ก่อนเริ่มพูดใหม่
	time.Sleep(100 * time.Millisecond)

	// สร้าง SpeechSynthesisUtterance object สำหรับข้อความที่ต้องการพูด
	utterance := js.Global.Get("SpeechSynthesisUtterance").New(message)
	utterance.Set("volume", volume) // ตั้งค่าความดังของเสียง

	// สั่งให้เบราว์เซอร์พูดข้อความ
	speechSynthesis.Call("speak", utterance)

	return nil // Go function ที่เรียกจาก JS ต้อง return interface{}
}
