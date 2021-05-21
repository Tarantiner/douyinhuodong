package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	At    string // 活动标题
	Ar    string // 活动要求
	Ss    string // 开始报名
	Es    string // 结束报名
	St    string // 开始推广
	Et    string // 结束推广
	Yj    string // 佣金率
	Fw    string // 服务费率
	Gj    string // 预估成交价
	Wx    string // 微信
	Phone string // 手机
	Ts    string // 任务开始
	Te    string // 任务结束
	Gp    string // 任务间隔
}

type Cookie struct {
	ShopType string `json:"buyin_shop_type"`
	AppID    string `json:"buyin_app_id"`
	SasID    string `json:"SASID"`
}

var mp = make(map[string]*Task)
var client = http.Client{}

func init() {
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		//fmt.Println(path)
		//楷体:simkai.ttf
		//黑体:simhei.ttf
		if strings.Contains(path, "simkai.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

func subTask(t *Task) int {
	// 读取cookie
	f, err := os.Open("./cookie.json")
	if err != nil {
		return 1
	}

	var ck Cookie
	dec := json.NewDecoder(f)
	err = dec.Decode(&ck)
	if err != nil {
		return 3
	}

	// 开始报名
	tmp, err := time.ParseInLocation("2006/01/02 15:04:05", t.Ss, time.Local)
	if err != nil {
		return 0
	}
	ast := tmp.Format("2006-01-02T15:04:05Z")
	ast1 := tmp.Format("2006-01-02")

	// 结束报名
	tmp, err = time.ParseInLocation("2006/01/02 15:04:05", t.Es, time.Local)
	if err != nil {
		return 0
	}
	aet := tmp.Format("2006-01-02T15:04:05Z")
	aet1 := tmp.Format("2006-01-02")

	// 开始推广
	tmp, err = time.ParseInLocation("2006/01/02 15:04:05", t.St, time.Local)
	if err != nil {
		return 0
	}
	pst := tmp.Format("2006-01-02T15:04:05Z")
	pst1 := tmp.Format("2006-01-02")

	// 结束推广
	tmp, err = time.ParseInLocation("2006/01/02 15:04:05", t.Et, time.Local)
	if err != nil {
		return 0
	}
	pet := tmp.Format("2006-01-02T15:04:05Z")
	pet1 := tmp.Format("2006-01-02")

	m := map[string]interface{}{"activity_kind": 0, "activity_type": 1, "threshold": false, "activity_name": t.At, "activity_desc": t.Ar,
		"apply_time": []interface{}{ast, aet},
		//"apply_time":      []interface{}{"2021-05-20T13:05:24.569Z", "2021-05-23T13:05:24.569Z"},
		"promote_time": []interface{}{pst, pet},
		//"promote_time":    []interface{}{"2021-05-21T13:05:29.637Z", "2021-05-24T13:05:29.637Z"},
		"commission_rate": t.Yj, "service_rate": t.Fw, "estimated_single_sale": t.Gj, "wechat_id": t.Wx, "phone_num": t.Phone,
		"category": []interface{}{4, 16, 18, 19, 20, 9, 5, 13, 8, 11, 10, 2, 15, 6, 14, 7, 3, 17, 12}, "online": true,
		"promote_start_time": pst1, "promote_end_time": pet1, "apply_start_time": ast1, "apply_end_time": aet1,
		"specified_shop_ids": ""}
	b, err := json.Marshal(m)
	if err != nil {
		return 0
	}

	req, err := http.NewRequest("POST", "https://buyin.jinritemai.com/api/institution/activity/edit", bytes.NewReader(b))
	if err != nil {
		return 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://buyin.jinritemai.com/dashboard/institution/activity")
	req.Header.Set("Cookie", fmt.Sprintf("buyin_shop_type=%s; buyin_app_id=%s; SASID=%s", ck.ShopType, ck.AppID, ck.SasID))

	rsp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer rsp.Body.Close()
	b, err = ioutil.ReadAll(rsp.Body)
	fmt.Println(strings.Repeat("#", 80))
	fmt.Println(string(b))
	if strings.Contains(string(b), "success") {
		log.Println("成功发布任务", t.At)
		return 0
	}
	if strings.Contains(string(b), "用户未登录") {
		log.Println("发布失败，用户未登录")
		return 1
	}
	if strings.Contains(string(b), "活动最多") {
		log.Println("发布失败，超过发布限制")
		return 2
	}
	log.Println("发布失败，接口返回异常")
	return 3
}

func main() {
	a := app.New()
	w := a.NewWindow("活动发布")

	title := widget.NewLabel("活动标题：")
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("少于30字符")
	titleEntry.Move(fyne.NewPos(80, 0))
	titleEntry.Resize(fyne.NewSize(400, 37))

	requirements := widget.NewLabel("活动要求：")
	requireEntry := widget.NewMultiLineEntry()
	requireEntry.SetPlaceHolder("少于200字符")

	// 报名
	startSign := widget.NewLabel("开始报名：")
	startSignEntry := widget.NewEntry()
	startSignEntry.Move(fyne.NewPos(80, 0))
	startSignEntry.Resize(fyne.NewSize(400, 37))

	endSign := widget.NewLabel("结束报名：")
	esEntry := widget.NewEntry()
	esEntry.Move(fyne.NewPos(80, 0))
	esEntry.Resize(fyne.NewSize(400, 37))

	t1 := time.Now().AddDate(0, 0, 1).Format("2006/01/02 15:04:05")
	tm1 := strings.Split(t1, " ")[0] + " 00:00:00"
	startSignEntry.SetText(tm1)

	t2 := time.Now().AddDate(0, 0, 2).Format("2006/01/02 15:04:05")
	tm2 := strings.Split(t2, " ")[0] + " 00:00:00"
	esEntry.SetText(tm2)

	// 推广
	startTg := widget.NewLabel("开始推广：")
	startTgEntry := widget.NewEntry()
	startTgEntry.Move(fyne.NewPos(80, 0))
	startTgEntry.Resize(fyne.NewSize(400, 37))

	endTg := widget.NewLabel("结束推广：")
	etEntry := widget.NewEntry()
	etEntry.Move(fyne.NewPos(80, 0))
	etEntry.Resize(fyne.NewSize(400, 37))

	t3 := time.Now().AddDate(0, 0, 2).Format("2006/01/02 15:04:05")
	tm3 := strings.Split(t3, " ")[0] + " 00:00:00"
	startTgEntry.SetText(tm3)

	t4 := time.Now().AddDate(0, 0, 3).Format("2006/01/02 15:04:05")
	tm4 := strings.Split(t4, " ")[0] + " 00:00:00"
	etEntry.SetText(tm4)

	// 佣金率
	yjLabel := widget.NewLabel("佣金率：")
	yjEntry := widget.NewEntry()
	yjEntry.Move(fyne.NewPos(80, 0))
	yjEntry.Resize(fyne.NewSize(400, 37))
	yjEntry.SetPlaceHolder("1-50")

	// 服务费率
	fwLabel := widget.NewLabel("服务费率：")
	fwEntry := widget.NewEntry()
	fwEntry.Move(fyne.NewPos(80, 0))
	fwEntry.Resize(fyne.NewSize(400, 37))
	fwEntry.SetPlaceHolder("佣金率+服务费率<=90")

	// 预估价
	gjLabel := widget.NewLabel("预估价：")
	gjEntry := widget.NewEntry()
	gjEntry.Move(fyne.NewPos(80, 0))
	gjEntry.Resize(fyne.NewSize(400, 37))
	gjEntry.SetText("9999")

	wxLabel := widget.NewLabel("联系微信：")
	wxEntry := widget.NewEntry()
	wxEntry.Move(fyne.NewPos(80, 0))
	wxEntry.Resize(fyne.NewSize(400, 37))

	phoneLabel := widget.NewLabel("联系手机：")
	phoneEntry := widget.NewEntry()
	phoneEntry.Move(fyne.NewPos(80, 0))
	phoneEntry.Resize(fyne.NewSize(400, 37))

	// 发布任务
	startT := widget.NewLabel("开始任务：")
	startTEntry := widget.NewEntry()
	startTEntry.Move(fyne.NewPos(80, 0))
	startTEntry.Resize(fyne.NewSize(400, 37))

	endT := widget.NewLabel("结束任务：")
	tEntry := widget.NewEntry()
	tEntry.Move(fyne.NewPos(80, 0))
	tEntry.Resize(fyne.NewSize(400, 37))

	gap := widget.NewLabel("间隔分钟：")
	gapEntry := widget.NewEntry()
	gapEntry.Move(fyne.NewPos(80, 0))
	gapEntry.Resize(fyne.NewSize(400, 37))

	t5 := time.Now().Format("2006/01/02 15:04:05")
	startTEntry.SetText(t5)

	t6 := time.Now().AddDate(0, 0, 1).Format("2006/01/02 15:04:05")
	tEntry.SetText(t6)

	gapEntry.SetText("8")

	infoLabel := widget.NewLabel("")

	var upTask = func() {
		at := titleEntry.Text     // 活动标题
		ar := requireEntry.Text   // 活动要求
		ss := startSignEntry.Text // 开始报名
		es := esEntry.Text        // 结束报名
		st := startTgEntry.Text   // 开始推广
		et := etEntry.Text        // 结束推广
		yj := yjEntry.Text        // 佣金率
		fw := fwEntry.Text        // 服务费率
		gj := gjEntry.Text        // 预估成交价
		wx := wxEntry.Text        // 微信
		phone := phoneEntry.Text  // 手机
		ts := startTEntry.Text    // 任务开始
		te := tEntry.Text         // 任务结束
		gp := gapEntry.Text       // 任务间隔

		// 任务检查
		if at == "" || ar == "" || ts == "" || te == "" {
			infoLabel.SetText("活动标题、活动要求、任务开始、任务结束是必填字段！")
			return
		}
		taskID := fmt.Sprintf("%x", md5.Sum([]byte(at+ar+ts+te)))
		if _, ok := mp[taskID]; ok {
			infoLabel.SetText("重复添加，【活动标题、活动要求、任务开始、任务结束】确定唯一任务!")
			return
		} else {
			infoLabel.SetText("")
		}

		itime, err := strconv.Atoi(gp)
		if err != nil {
			infoLabel.SetText("无效任务间隔！")
			return
		}
		itime = itime * 60

		if time.Now().Unix() > 1624151797 { // 程序有效期到2021/06/20 9：16
			infoLabel.SetText("发布任务异常！")
			return
		}

		// 添加任务
		t := &Task{at, ar, ss, es, st, et, yj, fw, gj, wx, phone, ts, te, gp}
		mp[taskID] = t

		infoLabel.SetText(fmt.Sprintf("当前%d个任务进行中", len(mp)))
		time.Sleep(time.Second * 2)
		infoLabel.SetText("")

		// 执行任务
		go func() {
			for {
				if _, ok := mp[taskID]; ok {
					// 判断时间是否过期
					now := time.Now().Unix()
					startTime, err := time.ParseInLocation("2006/01/02 15:04:05", ts, time.Local)
					if err != nil {
						delete(mp, taskID)
						break
					}
					_s := startTime.Unix()
					endTime, err := time.ParseInLocation("2006/01/02 15:04:05", te, time.Local)
					if err != nil {
						delete(mp, taskID)
						break
					}
					_e := endTime.Unix()
					if now >= _s && now <= _e {
						status := subTask(t)
						if status == 1 { // cookie过期
							infoLabel.SetText("请更新cookie重新添加任务")
							delete(mp, taskID)
							break
						}
						if status == 2 { // 超过限制
							infoLabel.SetText("超过活动数量限制，无法继续添加任务")
							delete(mp, taskID)
							break
						}
						if status == 3 { // 其他原因无法加入任务
							infoLabel.SetText("触发其他规则导致无法加入任务！")
							delete(mp, taskID)
							break
						}
					} else if now > _e { // 超过结束发布时间
						fmt.Println("超过时间，将结束任务")
						break
					}
					time.Sleep(time.Second * time.Duration(itime))
				} else {
					fmt.Println("一个任务被取消")
					break
				}
			}
		}()
	}

	upBtn := widget.NewButton("提交任务", upTask)
	upBtn.Resize(fyne.NewSize(100, 36))
	upBtn.Move(fyne.NewPos(210, 20))

	w.SetContent(container.NewVBox(container.NewWithoutLayout(title, titleEntry), requirements, requireEntry,
		container.NewWithoutLayout(startSign, startSignEntry),
		container.NewWithoutLayout(endSign, esEntry),
		container.NewWithoutLayout(startTg, startTgEntry),
		container.NewWithoutLayout(endTg, etEntry),
		container.NewWithoutLayout(yjLabel, yjEntry),
		container.NewWithoutLayout(fwLabel, fwEntry),
		container.NewWithoutLayout(gjLabel, gjEntry),
		container.NewWithoutLayout(wxLabel, wxEntry),
		container.NewWithoutLayout(phoneLabel, phoneEntry),
		container.NewWithoutLayout(startT, startTEntry),
		container.NewWithoutLayout(endT, tEntry),
		container.NewWithoutLayout(gap, gapEntry),
		infoLabel,
		container.NewWithoutLayout(upBtn)))

	// 任务管理
	manager := fyne.NewMenuItem("任务管理", func() {
		go func() {
			w := a.NewWindow("任务中心")
			clis := container.New(layout.NewVBoxLayout())
			for taskID := range mp {
				t := mp[taskID]
				l1 := widget.NewLabel(fmt.Sprintf("活动标题：%s", t.At))
				l2 := widget.NewLabel(fmt.Sprintf("活动要求：%s", t.Ar))
				l3 := widget.NewLabel(fmt.Sprintf("开始报名：%s", t.Ss))
				l4 := widget.NewLabel(fmt.Sprintf("结束报名：%s", t.Es))
				l5 := widget.NewLabel(fmt.Sprintf("开始推广：%s", t.St))
				l6 := widget.NewLabel(fmt.Sprintf("结束推广：%s", t.Et))

				la1 := widget.NewLabel(fmt.Sprintf("佣金率：%s", t.Yj))
				la2 := widget.NewLabel(fmt.Sprintf("服务费率：%s", t.Fw))
				la3 := widget.NewLabel(fmt.Sprintf("预估成交价：%s", t.Gj))

				l7 := widget.NewLabel(fmt.Sprintf("微信：%s", t.Wx))
				l8 := widget.NewLabel(fmt.Sprintf("手机：%s", t.Phone))
				l9 := widget.NewLabel(fmt.Sprintf("任务开始：%s", t.Ts))
				l10 := widget.NewLabel(fmt.Sprintf("任务结束：%s", t.Te))
				l11 := widget.NewLabel(fmt.Sprintf("间隔时间：%s", t.Gp))
				btn := widget.NewButton("取消任务", func() {
					delete(mp, taskID)
					w.Close()
				})
				c := container.New(layout.NewVBoxLayout(), l1, l2, l3, l4, l5, l6, la1, la2, la3, l7, l8, l9, l10, l11, btn)
				clis.Add(c)
			}
			w.SetContent(container.NewScroll(clis))
			w.Resize(fyne.NewSize(450, 600))
			w.Show()
		}()
	})
	mu := fyne.NewMainMenu(fyne.NewMenu("管理", manager))
	w.SetMainMenu(mu)

	w.Resize(fyne.NewSize(520, 800))
	w.ShowAndRun()
	os.Unsetenv("FYNE_FONT")
}
