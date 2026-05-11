package ai

import "testing"

func TestAIChat_GenerateGreeting(t *testing.T) {
	chat := NewAIChat(1, "张三", "豪爽的东北大哥", "你瞅啥？")
	greeting := chat.GenerateGreeting()
	expected := "张三 来啦！你瞅啥？"
	if greeting != expected {
		t.Errorf("greeting = %q, want %q", greeting, expected)
	}
}

func TestAIChat_GenerateGreeting_NoCatchphrase(t *testing.T) {
	chat := NewAIChat(2, "李四", "沉默寡言的技术宅", "")
	greeting := chat.GenerateGreeting()
	expected := "李四 来了，开始吧！"
	if greeting != expected {
		t.Errorf("greeting = %q, want %q", greeting, expected)
	}
}

func TestAIChat_GenerateCatchphrase(t *testing.T) {
	chat := NewAIChat(1, "张三", "豪爽的东北大哥", "你瞅啥？")
	phrase := chat.GenerateCatchphrase()
	if phrase != "你瞅啥？" {
		t.Errorf("catchphrase = %q, want %q", phrase, "你瞅啥？")
	}
}

func TestAIChat_GenerateCatchphrase_Fallback(t *testing.T) {
	chat := NewAIChat(2, "李四", "沉默寡言的技术宅", "")
	phrase := chat.GenerateCatchphrase()
	if phrase != "嘿嘿嘿" {
		t.Errorf("catchphrase = %q, want %q", phrase, "嘿嘿嘿")
	}
}

func TestAIChat_GeneratePlayComment(t *testing.T) {
	chat := NewAIChat(1, "张三", "豪爽的东北大哥", "你瞅啥？")
	comment := chat.GeneratePlayComment("play", []int{0, 13, 26})
	if comment != "接招！" {
		t.Errorf("play comment = %q, want %q", comment, "接招！")
	}
}

func TestAIChat_GeneratePlayComment_Pass(t *testing.T) {
	chat := NewAIChat(1, "张三", "豪爽的东北大哥", "你瞅啥？")
	comment := chat.GeneratePlayComment("pass", []int{})
	if comment != "过过过" {
		t.Errorf("pass comment = %q, want %q", comment, "过过过")
	}
}
