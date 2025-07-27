// main.go
package main

import (
	"fmt"
	"syscall/js"
	"time"
)

func main() {
	c := make(chan struct{}, 0)
	fmt.Println("Go WebAssembly Stupid Alarm Initialized!")

	js.Global().Set("speakAlarm", js.FuncOf(speakAlarmFunc))

	<-c
}

// getSpeechSynthesis พยายามเข้าถึง speechSynthesis object และรอจนกว่าจะพร้อม
func getSpeechSynthesis() js.Value {
	speechSynthesis := js.Global().Get("speechSynthesis")
	if !speechSynthesis.Truthy() {
		return js.Undefined()
	}

	// SpeechSynthesis API ต้องการเวลาในการโหลดเสียง (voices) ให้พร้อมใช้งาน
	// เราจะวนรอสักพัก หากเสียงยังไม่พร้อม
	for i := 0; i < 50; i++ { // ลอง 50 ครั้ง (รวม 500ms)
		// ตรวจสอบว่ามีเสียง (voices) ที่สามารถใช้งานได้หรือไม่
		if js.Global().Get("speechSynthesis").Call("getVoices").Length() > 0 {
			return speechSynthesis
		}
		time.Sleep(10 * time.Millisecond) // รอ 10 มิลลิวินาที
	}
	return js.Undefined()
}

// speakAlarmFunc เป็นฟังก์ชันที่จะถูกเรียกจาก JavaScript
func speakAlarmFunc(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		fmt.Println("speakAlarm: No message provided.")
		return nil
	}
	message := args[0].String()
	volume := 1.0

	if len(args) > 1 && args[1].Type() == js.TypeNumber {
		volume = args[1].Float()
		if volume < 0.0 {
			volume = 0.0
		} else if volume > 1.0 {
			volume = 1.0
		}
	}

	fmt.Printf("Attempting to speak: \"%s\" with volume %.2f\n", message, volume)

	speechSynthesis := getSpeechSynthesis()
	if !speechSynthesis.Truthy() {
		fmt.Println("Web Speech API not available or not ready.")
		return nil
	}

	if speechSynthesis.Get("speaking").Bool() {
		fmt.Println("SpeechSynthesis is currently speaking, canceling previous utterance.")
		speechSynthesis.Call("cancel")
		time.Sleep(100 * time.Millisecond)
	}

	utterance := js.Global().Get("SpeechSynthesisUtterance").New(message)

	// *** การแก้ไข: กำหนดภาษาให้ชัดเจน ***
	utterance.Set("lang", "th-TH") // กำหนดภาษาไทย
	// คุณสามารถลองเปลี่ยนเป็น "en-US" สำหรับภาษาอังกฤษได้ถ้าต้องการ

	utterance.Set("volume", volume)

	speechSynthesis.Call("speak", utterance)
	fmt.Printf("Successfully triggered speech for: \"%s\"\n", message)

	return nil
}
