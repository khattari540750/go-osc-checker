package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/hypebeast/go-osc/osc"
	"gopkg.in/yaml.v3"
)

// AppConfig アプリケーション全体の設定
type AppConfig struct {
	App      AppSettings      `yaml:"app"`
	Sender   SenderSettings   `yaml:"sender"`
	Receiver ReceiverSettings `yaml:"receiver"`
}

// AppSettings アプリケーション基本設定
type AppSettings struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// WindowSettings ウィンドウ設定
type WindowSettings struct {
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
	Title  string `yaml:"title"`
}

// SenderArgument 送信側の引数定義
type SenderArgument struct {
	Type         string `yaml:"type"`
	DefaultValue string `yaml:"default_value"`
	Description  string `yaml:"description"`
}

// SenderTarget 送信先設定
type SenderTarget struct {
	Name      string           `yaml:"name"`
	Host      string           `yaml:"host"`
	Port      int              `yaml:"port"`
	Address   string           `yaml:"address"`
	Arguments []SenderArgument `yaml:"arguments"`
}

// SenderSettings 送信側設定
type SenderSettings struct {
	List   []SenderTarget `yaml:"list"`
	Window WindowSettings `yaml:"window"`
}

// ReceiverSettings 受信側設定
type ReceiverSettings struct {
	DefaultPort   int            `yaml:"default_port"`
	Window        WindowSettings `yaml:"window"`
	MaxLogEntries int            `yaml:"max_log_entries"`
}

// OSCArgument OSC引数の構造体
type OSCArgument struct {
	Type  string // "int", "float", "string", "bool"
	Value string
}

// OSCMessage 受信したOSCメッセージ
type OSCMessage struct {
	Timestamp string
	Address   string
	Values    string
}

// LoadConfig 設定ファイルを読み込む
func LoadConfig(filename string) (*AppConfig, error) {
	config := &AppConfig{}

	// ファイルが存在しない場合はデフォルト設定を作成
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("設定ファイル %s が見つかりません。デフォルト設定を使用します。", filename)
		return getDefaultConfig(), nil
	}

	// ファイルを読み込み
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// YAMLをパース
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	log.Printf("設定ファイル %s を読み込みました", filename)
	return config, nil
}

// getDefaultConfig デフォルト設定を返す
func getDefaultConfig() *AppConfig {
	return &AppConfig{
		App: AppSettings{
			Name:    "OSC Checker",
			Version: "1.0.0",
		},
		Sender: SenderSettings{
			List: []SenderTarget{
				{
					Name:    "Default",
					Host:    "127.0.0.1",
					Port:    7000,
					Address: "/test",
					Arguments: []SenderArgument{
						{
							Type:         "int",
							DefaultValue: "42",
							Description:  "Test integer",
						},
					},
				},
			},
			Window: WindowSettings{
				Width:  900,
				Height: 600,
				Title:  "OSC Sender",
			},
		},
		Receiver: ReceiverSettings{
			DefaultPort: 7000,
			Window: WindowSettings{
				Width:  1000,
				Height: 700,
				Title:  "OSC Receiver",
			},
			MaxLogEntries: 100,
		},
	}
}

// createSenderSection 単一の送信セクションを作成
func createSenderSection(target SenderTarget, index int, updateHistory func(string)) *widget.Card {
	// OSC送信用のUI要素（固定サイズコンテナでラップ）
	hostEntry := widget.NewEntry()
	hostEntry.SetText(target.Host)
	hostEntry.SetPlaceHolder("Host IP")
	hostEntry.Resize(fyne.NewSize(120, 32))
	hostContainer := container.NewWithoutLayout(hostEntry)
	hostContainer.Resize(fyne.NewSize(120, 32))
	hostEntry.Move(fyne.NewPos(0, 0))

	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", target.Port))
	portEntry.SetPlaceHolder("Port")
	portEntry.Resize(fyne.NewSize(80, 32))
	portContainer := container.NewWithoutLayout(portEntry)
	portContainer.Resize(fyne.NewSize(80, 32))
	portEntry.Move(fyne.NewPos(0, 0))

	addressEntry := widget.NewEntry()
	addressEntry.SetText(target.Address)
	addressEntry.SetPlaceHolder("OSC Address")
	addressEntry.Resize(fyne.NewSize(200, 32))
	addressContainer := container.NewWithoutLayout(addressEntry)
	addressContainer.Resize(fyne.NewSize(200, 32))
	addressEntry.Move(fyne.NewPos(0, 0))

	// 設定ファイルから引数の初期値を読み込み
	var arguments []OSCArgument
	for _, argDef := range target.Arguments {
		arguments = append(arguments, OSCArgument{
			Type:  argDef.Type,
			Value: argDef.DefaultValue,
		})
	}

	argumentsContainer := container.NewVBox()

	// 引数表示を更新する関数
	var updateArgumentsDisplay func()
	updateArgumentsDisplay = func() {
		argumentsContainer.RemoveAll()
		for j, arg := range arguments {
			argIndex := j // クロージャ用

			// 引数の説明を取得
			description := ""
			if argIndex < len(target.Arguments) {
				description = target.Arguments[argIndex].Description
			}

			// 引数タイプ選択
			typeSelect := widget.NewSelect([]string{"int", "float", "string", "bool"}, func(value string) {
				if argIndex < len(arguments) {
					arguments[argIndex].Type = value
				}
			})
			typeSelect.SetSelected(arg.Type)

			// 引数値入力
			valueEntry := widget.NewEntry()
			valueEntry.SetText(arg.Value)
			valueEntry.OnChanged = func(value string) {
				if argIndex < len(arguments) {
					arguments[argIndex].Value = value
				}
			}

			// 削除ボタン
			removeBtn := widget.NewButton("✕", func() {
				if argIndex < len(arguments) {
					arguments = append(arguments[:argIndex], arguments[argIndex+1:]...)
					updateArgumentsDisplay()
				}
			})

			// 引数行の構成
			labelText := fmt.Sprintf("Arg%d:", j+1)
			if description != "" {
				labelText += fmt.Sprintf(" (%s)", description)
			}

			argRow := container.NewHBox(
				widget.NewLabel(labelText),
				typeSelect,
				valueEntry,
				removeBtn,
			)
			argumentsContainer.Add(argRow)
		}
		argumentsContainer.Refresh()
	}

	// 初期表示
	updateArgumentsDisplay()

	// 引数追加ボタン
	addArgBtn := widget.NewButton("＋", func() {
		arguments = append(arguments, OSCArgument{Type: "int", Value: "0"})
		updateArgumentsDisplay()
	})

	// 送信ボタン
	sendBtn := widget.NewButton("Send", func() {
		host := hostEntry.Text
		portStr := portEntry.Text
		address := addressEntry.Text

		if host == "" || portStr == "" || address == "" {
			log.Printf("送信エラー [%s]: ホスト、ポート、アドレスを入力してください", target.Name)
			return
		}

		// ポート番号をパース
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("ポート番号が無効です [%s]: %s", target.Name, portStr)
			return
		}

		// OSCクライアントを作成
		client := osc.NewClient(host, port)

		// OSCメッセージを作成
		msg := osc.NewMessage(address)

		// 引数を追加
		for _, arg := range arguments {
			switch arg.Type {
			case "int":
				if val, err := strconv.Atoi(arg.Value); err == nil {
					msg.Append(int32(val))
				} else {
					log.Printf("int変換エラー [%s]: %s", target.Name, arg.Value)
					return
				}
			case "float":
				if val, err := strconv.ParseFloat(arg.Value, 32); err == nil {
					msg.Append(float32(val))
				} else {
					log.Printf("float変換エラー [%s]: %s", target.Name, arg.Value)
					return
				}
			case "string":
				msg.Append(arg.Value)
			case "bool":
				if val, err := strconv.ParseBool(arg.Value); err == nil {
					msg.Append(val)
				} else {
					log.Printf("bool変換エラー [%s]: %s", target.Name, arg.Value)
					return
				}
			}
		}

		// OSCメッセージを送信
		err = client.Send(msg)
		if err != nil {
			log.Printf("OSC送信エラー [%s]: %v", target.Name, err)
			return
		}

		// 引数の情報をログ出力
		var argInfo []string
		for _, arg := range arguments {
			argInfo = append(argInfo, fmt.Sprintf("%s:%s", arg.Type, arg.Value))
		}

		logMsg := fmt.Sprintf("OSC送信完了 [%s]: %s:%d %s [%s]", target.Name, host, port, address, strings.Join(argInfo, ", "))
		log.Printf(logMsg)
		
		// 送信履歴を更新
		timestamp := time.Now().Format("15:04:05")
		historyMsg := fmt.Sprintf("%s | %s → %s:%d %s [%s]", timestamp, target.Name, host, port, address, strings.Join(argInfo, ", "))
		updateHistory(historyMsg)
	})
	
	// Sendボタンのサイズを大きく設定
	sendBtn.Resize(fyne.NewSize(80, 40))
	
	// nameラベルを大きなフォントで作成
	nameLabel := widget.NewRichTextFromMarkdown(fmt.Sprintf("## %s", target.Name))
	nameLabel.Wrapping = fyne.TextWrapOff

	// セクションのレイアウト
	sectionContent := container.NewVBox(
		// nameとSendボタンを横並び
		container.NewHBox(
			sendBtn,
			nameLabel,
			layout.NewSpacer(),
		),
		
		widget.NewSeparator(),
		
		// 送信先設定とアドレス - 固定位置レイアウト
		func() *fyne.Container {
			// 位置計算用の定数
			const (
				ipLabelWidth    = 30  // "IP:" の幅
				hostFieldWidth  = 120 // IPフィールドの幅
				portLabelWidth  = 50  // "Port:" の幅
				portFieldWidth  = 80  // Portフィールドの幅
				addrLabelWidth  = 80  // "OSC Addr:" の幅
				addrFieldWidth  = 200 // OSC Addrフィールドの幅
				spacing         = 5   // 要素間のスペース
			)

			// ラベルを作成
			ipLabel := widget.NewLabel("IP:")
			portLabel := widget.NewLabel("Port:")
			addrLabel := widget.NewLabel("OSC Addr:")

			// WithoutLayoutコンテナを作成
			layoutContainer := container.NewWithoutLayout()
			layoutContainer.Resize(fyne.NewSize(600, 32))

			// 各要素の位置を計算して配置
			xPos := float32(0)

			// IP: ラベル
			ipLabel.Move(fyne.NewPos(xPos, 0))
			ipLabel.Resize(fyne.NewSize(ipLabelWidth, 32))
			layoutContainer.Add(ipLabel)
			xPos += ipLabelWidth + spacing

			// IPフィールド
			hostContainer.Move(fyne.NewPos(xPos, 0))
			layoutContainer.Add(hostContainer)
			xPos += hostFieldWidth + spacing

			// Port: ラベル
			portLabel.Move(fyne.NewPos(xPos, 0))
			portLabel.Resize(fyne.NewSize(portLabelWidth, 32))
			layoutContainer.Add(portLabel)
			xPos += portLabelWidth + spacing

			// Portフィールド
			portContainer.Move(fyne.NewPos(xPos, 0))
			layoutContainer.Add(portContainer)
			xPos += portFieldWidth + spacing

			// OSC Addr: ラベル
			addrLabel.Move(fyne.NewPos(xPos, 0))
			addrLabel.Resize(fyne.NewSize(addrLabelWidth, 32))
			layoutContainer.Add(addrLabel)
			xPos += addrLabelWidth + spacing

			// OSC Addrフィールド
			addressContainer.Move(fyne.NewPos(xPos, 0))
			layoutContainer.Add(addressContainer)

			return layoutContainer
		}(),

		widget.NewSeparator(),

		// 引数設定
		container.NewHBox(
			addArgBtn,
		),

		argumentsContainer,
	)

	return widget.NewCard(
		"",
		"",
		sectionContent,
	)
}

func main() {
	// 設定ファイルを読み込み
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗しました: %v", err)
	}

	a := app.NewWithID("com.example.oscchecker")

	// メッセージ管理用のスライス
	var messages []OSCMessage
	
	// 送信履歴管理用のスライス
	var sendHistory []string

	// Senderウィンドウ
	senderWin := a.NewWindow(config.Sender.Window.Title)

	// 送信履歴（簡易版）
	historyLabel := widget.NewLabel("Send history will be displayed here")
	historyScroll := container.NewScroll(historyLabel)
	historyScroll.SetMinSize(fyne.NewSize(0, 150))

	// 送信履歴を更新する関数
	updateSendHistory := func(msg string) {
		sendHistory = append([]string{msg}, sendHistory...)
		if len(sendHistory) > 50 { // 最大50件まで保持
			sendHistory = sendHistory[:50]
		}
		
		historyText := strings.Join(sendHistory, "\n")
		if historyText == "" {
			historyText = "Send history will be displayed here"
		}
		historyLabel.SetText(historyText)
	}

	// 複数送信セクションを格納するコンテナ
	sendersContainer := container.NewVBox()

	// 各送信先に対してUIセクションを作成
	for i, target := range config.Sender.List {
		sectionCard := createSenderSection(target, i, updateSendHistory)
		sendersContainer.Add(sectionCard)
		
		// 最後以外はセパレータを追加
		if i < len(config.Sender.List)-1 {
			sendersContainer.Add(widget.NewSeparator())
		}
	}

	// メインレイアウト
	senderContent := container.NewBorder(
		widget.NewCard("OSC Sender", "", nil), // top
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabel("Send History:"),
			historyScroll,
		), // bottom
		nil, // left
		nil, // right
		container.NewScroll(sendersContainer), // center - スクロール可能
	)

	senderWin.SetContent(senderContent)
	senderWin.Resize(fyne.NewSize(float32(config.Sender.Window.Width), float32(config.Sender.Window.Height)))
	senderWin.Show()

	// Receiverウィンドウ
	receiverWin := a.NewWindow(config.Receiver.Window.Title)

	// OSC受信用のUI要素
	receiverPortEntry := widget.NewEntry()
	receiverPortEntry.SetText(fmt.Sprintf("%d", config.Receiver.DefaultPort))
	receiverPortEntry.SetPlaceHolder("Port Number")
	receiverPortEntry.Resize(fyne.NewSize(80, 32))

	statusLabel := widget.NewLabel("Stopped")
	statusLabel.Importance = widget.MediumImportance

	// Address Filter Entry
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Address Filter (e.g. /test*, /osc/*, empty=all)")

	// メッセージログ用のリスト
	logContent := widget.NewLabel("Message log will be displayed here")
	logScroll := container.NewScroll(logContent)

	// 受信メッセージカウンタ
	messageCountLabel := widget.NewLabel("Received: 0")

	// ログコンテンツを更新する関数
	updateLogContent := func() {
		var logText string
		filter := filterEntry.Text

		for _, msg := range messages {
			// フィルター機能
			shouldShow := false
			if filter == "" {
				shouldShow = true
			} else if strings.HasSuffix(filter, "*") {
				prefix := strings.TrimSuffix(filter, "*")
				shouldShow = strings.HasPrefix(msg.Address, prefix)
			} else {
				shouldShow = strings.Contains(msg.Address, filter)
			}

			if shouldShow {
				logText += fmt.Sprintf("%s | %s | %s\n", msg.Timestamp, msg.Address, msg.Values)
			}
		}
		if logText == "" {
			logText = "Message log will be displayed here"
		}
		logContent.SetText(logText)
	}

	// フィルター入力が変更されたらリアルタイムで表示を更新
	filterEntry.OnChanged = func(content string) {
		updateLogContent()
	}

	// メッセージ追加関数
	addMessage := func(address, values string) {
		timestamp := time.Now().Format("15:04:05")
		newMsg := OSCMessage{
			Timestamp: timestamp,
			Address:   address,
			Values:    values,
		}

		messages = append([]OSCMessage{newMsg}, messages...)
		if len(messages) > config.Receiver.MaxLogEntries {
			messages = messages[:config.Receiver.MaxLogEntries]
		}

		messageCountLabel.SetText(fmt.Sprintf("Received: %d", len(messages)))
		updateLogContent()
	}

	// 受信制御用の変数
	var startStopBtn *widget.Button
	var oscServer *osc.Server
	var isReceiving bool

	// Start/Stopボタン
	startStopBtn = widget.NewButton("Start", func() {
		if !isReceiving {
			// スタート時にログをクリア
			messages = []OSCMessage{}
			messageCountLabel.SetText("Received: 0")
			updateLogContent()

			// OSC受信開始
			port, err := strconv.Atoi(receiverPortEntry.Text)
			if err != nil {
				log.Printf("ポート番号が無効です: %s", receiverPortEntry.Text)
				return
			}

			addr := fmt.Sprintf("127.0.0.1:%d", port)

			// ディスパッチャーを作成
			dispatcher := osc.NewStandardDispatcher()

			// すべてのメッセージを受信するハンドラを追加
			dispatcher.AddMsgHandler("*", func(msg *osc.Message) {
				address := msg.Address

				// 引数を文字列に変換
				var values []string
				for _, arg := range msg.Arguments {
					values = append(values, fmt.Sprintf("%v", arg))
				}
				valuesStr := strings.Join(values, ", ")

				// UIスレッドで更新
				addMessage(address, valuesStr)
				log.Printf("OSC受信: %s [%s]", address, valuesStr)
			})

			// サーバーを作成
			oscServer = &osc.Server{
				Addr:       addr,
				Dispatcher: dispatcher,
			}

			// サーバー開始
			go func() {
				err := oscServer.ListenAndServe()
				if err != nil {
					log.Printf("OSC受信エラー: %v", err)
				}
			}()

			startStopBtn.SetText("Stop")
			statusLabel.SetText("Receiving...")
			statusLabel.Importance = widget.SuccessImportance
			isReceiving = true
			log.Printf("OSC受信を開始 (ポート: %d)", port)
		} else {
			// OSC受信停止
			if oscServer != nil {
				oscServer = nil
			}
			startStopBtn.SetText("Start")
			statusLabel.SetText("Stopped")
			statusLabel.Importance = widget.MediumImportance
			isReceiving = false
			log.Println("OSC受信を停止")
		}
		statusLabel.Refresh()
	})

	startStopBtn.Resize(fyne.NewSize(150, 50))

	// クリアボタン
	clearBtn := widget.NewButton("Clear", func() {
		messages = []OSCMessage{}
		messageCountLabel.SetText("Received: 0")
		updateLogContent()
	})

	// Receiverレイアウト構成
	receiverTopSection := container.NewVBox(
		widget.NewCard("OSC Receiver", "", nil),
		widget.NewSeparator(),

		// Connection Settings
		container.NewVBox(
			widget.NewLabelWithStyle("Connection Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(
				widget.NewLabel("Port:"),
				container.NewWithoutLayout(receiverPortEntry),
				layout.NewSpacer(),
			),
		),

		widget.NewSeparator(),

		// Start Button with Status
		container.NewHBox(
			startStopBtn,
			widget.NewSeparator(),
			statusLabel,
			layout.NewSpacer(),
		),

		widget.NewSeparator(),

		// Address Filter
		container.NewVBox(
			widget.NewLabel("Address Filter:"),
			filterEntry,
		),

		container.NewHBox(
			messageCountLabel,
		),

		widget.NewSeparator(),

		// Message Log Header with Clear Button
		container.NewHBox(
			widget.NewLabelWithStyle("Message Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			clearBtn,
			layout.NewSpacer(),
		),
	)

	// Receiverメイン画面
	receiverContent := container.NewBorder(
		receiverTopSection, // top
		nil,                // bottom
		nil,                // left
		nil,                // right
		logScroll,          // center
	)

	receiverWin.SetContent(receiverContent)
	receiverWin.Resize(fyne.NewSize(float32(config.Receiver.Window.Width), float32(config.Receiver.Window.Height)))
	receiverWin.Show()

	a.Run()
}
