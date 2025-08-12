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

// SenderSettings 送信側設定
type SenderSettings struct {
	DefaultHost    string         `yaml:"default_host"`
	DefaultPort    int            `yaml:"default_port"`
	DefaultAddress string         `yaml:"default_address"`
	Window         WindowSettings `yaml:"window"`
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
			DefaultHost:    "127.0.0.1",
			DefaultPort:    7000,
			DefaultAddress: "/test",
			Window: WindowSettings{
				Width:  400,
				Height: 350,
				Title:  "OSC Sender",
			},
		},
		Receiver: ReceiverSettings{
			DefaultPort: 7000,
			Window: WindowSettings{
				Width:  500,
				Height: 600,
				Title:  "OSC Receiver",
			},
			MaxLogEntries: 100,
		},
	}
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
	var messageCounter int

	// Senderウィンドウ
	senderWin := a.NewWindow(config.Sender.Window.Title)

	// OSC送信用のUI要素
	hostEntry := widget.NewEntry()
	hostEntry.SetText(config.Sender.DefaultHost)
	hostEntry.SetPlaceHolder("送信先IP")

	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", config.Sender.DefaultPort))
	portEntry.SetPlaceHolder("送信先ポート")

	addressEntry := widget.NewEntry()
	addressEntry.SetText(config.Sender.DefaultAddress)
	addressEntry.SetPlaceHolder("OSCアドレス (例: /test/sample)")

	// 引数管理用のスライス
	var arguments []OSCArgument

	// 引数リスト表示用
	argumentsContainer := container.NewVBox()

	// 引数を追加する関数（関数宣言を先に）
	var updateArgumentsDisplay func()

	updateArgumentsDisplay = func() {
		argumentsContainer.RemoveAll()
		for i, arg := range arguments {
			argIndex := i // クロージャ用

			// 引数タイプ選択
			typeSelect := widget.NewSelect([]string{"int", "float", "string", "bool"}, func(value string) {
				arguments[argIndex].Type = value
			})
			typeSelect.SetSelected(arg.Type)

			// 引数値入力
			valueEntry := widget.NewEntry()
			valueEntry.SetText(arg.Value)
			valueEntry.OnChanged = func(value string) {
				arguments[argIndex].Value = value
			}

			// 削除ボタン
			removeBtn := widget.NewButton("削除", func() {
				arguments = append(arguments[:argIndex], arguments[argIndex+1:]...)
				updateArgumentsDisplay()
			})

			argRow := container.NewHBox(
				widget.NewLabel(fmt.Sprintf("引数%d:", i+1)),
				typeSelect,
				valueEntry,
				removeBtn,
			)
			argumentsContainer.Add(argRow)
		}
		argumentsContainer.Refresh()
	}

	// 引数追加ボタン
	addArgBtn := widget.NewButton("引数追加", func() {
		arguments = append(arguments, OSCArgument{Type: "int", Value: "0"})
		updateArgumentsDisplay()
	})

	// 送信ボタン
	sendBtn := widget.NewButton("送信", func() {
		host := hostEntry.Text
		portStr := portEntry.Text
		address := addressEntry.Text

		if host == "" || portStr == "" || address == "" {
			log.Println("送信エラー: ホスト、ポート、アドレスを入力してください")
			return
		}

		// ポート番号をパース
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("ポート番号が無効です: %s", portStr)
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
					log.Printf("int変換エラー: %s", arg.Value)
					return
				}
			case "float":
				if val, err := strconv.ParseFloat(arg.Value, 32); err == nil {
					msg.Append(float32(val))
				} else {
					log.Printf("float変換エラー: %s", arg.Value)
					return
				}
			case "string":
				msg.Append(arg.Value)
			case "bool":
				if val, err := strconv.ParseBool(arg.Value); err == nil {
					msg.Append(val)
				} else {
					log.Printf("bool変換エラー: %s", arg.Value)
					return
				}
			}
		}

		// OSCメッセージを送信
		err = client.Send(msg)
		if err != nil {
			log.Printf("OSC送信エラー: %v", err)
			return
		}

		// 引数の情報をログ出力
		var argInfo []string
		for _, arg := range arguments {
			argInfo = append(argInfo, fmt.Sprintf("%s:%s", arg.Type, arg.Value))
		}

		log.Printf("OSC送信完了: %s:%d %s [%s]", host, port, address, strings.Join(argInfo, ", "))
	})

	// 送信履歴（簡易版）
	historyLabel := widget.NewLabel("送信履歴がここに表示されます")
	historyScroll := container.NewScroll(historyLabel)
	historyScroll.SetMinSize(fyne.NewSize(0, 100))

	// レイアウト構成
	topSection := container.NewVBox(
		widget.NewCard("OSC Sender", "送信設定", nil),

		// 送信先設定
		container.NewHBox(
			widget.NewLabel("送信先:"),
			hostEntry,
			widget.NewLabel(":"),
			portEntry,
		),

		// OSCアドレス
		container.NewHBox(
			widget.NewLabel("アドレス:"),
			addressEntry,
		),

		widget.NewSeparator(),

		// 引数設定
		container.NewHBox(
			widget.NewLabel("引数設定:"),
			addArgBtn,
		),

		argumentsContainer,

		widget.NewSeparator(),

		// 送信ボタン
		container.NewHBox(
			sendBtn,
		),

		widget.NewSeparator(),
		widget.NewLabel("送信履歴:"),
	)

	// メイン画面
	senderContent := container.NewBorder(
		topSection,    // top
		nil,           // bottom
		nil,           // left
		nil,           // right
		historyScroll, // center
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
	receiverPortEntry.Resize(fyne.NewSize(80, 32)) // 5桁のポート番号に適したサイズ

	statusLabel := widget.NewLabel("Stopped")
	statusLabel.Importance = widget.MediumImportance

	// Address Filter Entry
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Address Filter (e.g. /test*, /osc/*, empty=all)")

	// メッセージログ用のリスト（簡略化版）
	logContent := widget.NewLabel("Message log will be displayed here")
	logScroll := container.NewScroll(logContent)

	// 受信メッセージカウンタ
	messageCountLabel := widget.NewLabel("Received: 0")

	// OSC受信のモック関数（後でOSCライブラリに置き換え）
	updateLogContent := func() {
		var logText string
		filter := filterEntry.Text

		for _, msg := range messages {
			// フィルター機能：空の場合は全て表示、そうでなければマッチするもののみ
			shouldShow := false
			if filter == "" {
				shouldShow = true
			} else if strings.HasSuffix(filter, "*") {
				// ワイルドカード処理（例：/test* は /test で始まるアドレスにマッチ）
				prefix := strings.TrimSuffix(filter, "*")
				shouldShow = strings.HasPrefix(msg.Address, prefix)
			} else {
				// 完全一致または部分一致
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

		messageCounter++
		messageCountLabel.SetText(fmt.Sprintf("Received: %d", len(messages)))
		updateLogContent()
	}

	// 変数を先に宣言
	var startStopBtn *widget.Button
	var oscServer *osc.Server
	var isReceiving bool

	startStopBtn = widget.NewButton("Start", func() {
		if !isReceiving {
			// スタート時にログをクリア（クリアボタンと同じ挙動）
			messages = []OSCMessage{}
			messageCounter = 0
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
				// go-oscには明示的な停止メソッドがないため、サーバーを破棄
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

	// 開始ボタンを目立つサイズに設定
	startStopBtn.Resize(fyne.NewSize(150, 50))

	clearBtn := widget.NewButton("Clear", func() {
		messages = []OSCMessage{}
		messageCounter = 0
		messageCountLabel.SetText("Received: 0")
		updateLogContent()
	})

	// レイアウト構成
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

		// Start Button with Status (left-aligned, prominent)
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

	// メイン画面
	content := container.NewBorder(
		receiverTopSection, // top
		nil,                // bottom
		nil,                // left
		nil,                // right
		logScroll,          // center
	)

	receiverWin.SetContent(content)
	receiverWin.Resize(fyne.NewSize(float32(config.Receiver.Window.Width), float32(config.Receiver.Window.Height)))
	receiverWin.Show()

	a.Run()
}
