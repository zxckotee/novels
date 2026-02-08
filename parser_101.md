# Парсер *https://101kks.com/book/12544.html*

## Cloudflare / Turnstile (403 "Just a moment...")

`101kks.com` защищён Cloudflare и периодически отдаёт `403` страницу **"Just a moment..."** / Turnstile. В этом случае Playwright не может получить настоящую страницу книги/каталога, и `parser-service` вернёт `403 blocked by Cloudflare`.

Что делать:
- Обновить Playwright storage state файл `./cookies/101kks_storage.json` (в контейнере это `/data/101kks_storage.json`) так, чтобы там был актуальный `cf_clearance`
- Или передавать cookie прямо в запрос (полезно для быстрой проверки):

```bash
curl -sS -X POST http://localhost:8010/parse \
  -H 'Content-Type: application/json' \
  -d '{
    "url":"https://101kks.com/book/12544.html",
    "chapters_limit":1,
    "navigation_timeout_ms":300000,
    "cookie_header":"zh_choose=t; cf_clearance=YOUR_VALUE_HERE",
    "debug_http":true
  }'
```

curl 'https://101kks.com/book/12544.html' \
  -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \
  -H 'accept-language: ru,en;q=0.9' \
  -H 'cache-control: max-age=0' \
  -b 'zh_choose=t; cf_clearance=HCBZS3y57ehal5qQazG39l3Upvd1WekwKBJ0fwtm8k0-1769985973-1.2.1.1-.N8VCuX9TquQoNHyfnedDv.1KgONZM1dEQXc8pWRK9CgW19anSXcVzeKl.mdRJVaSfydMiHDB8wG4oBexjblPt.8WbLvF6BJFxD0ZPcVSYQd946.QYHSiRzoWcbQ0NStk355oSQVoeeBlRHX7llBeyOR1tcxe8d2xXiBgn1ZckpLmPzdi8S0zPCwr.kGsjGaZQM5jv98.X2W07eiG6_bK6Hulr3hqTavFgy3IVEjHFo; _ga=GA1.1.871789781.1769985973; _ga_DMRW3FNJ29=GS2.1.s1769985973$o1$g1$t1769985983$j50$l0$h0' \
  -H 'priority: u=0, i' \
  -H 'referer: https://101kks.com/booklist/detail/8.html' \
  -H 'sec-ch-ua: "Chromium";v="136", "YaBrowser";v="25.6", "Not.A/Brand";v="99", "Yowser";v="2.5"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: document' \
  -H 'sec-fetch-mode: navigate' \
  -H 'sec-fetch-site: same-origin' \
  -H 'sec-fetch-user: ?1' \
  -H 'upgrade-insecure-requests: 1' \
  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 YaBrowser/25.6.0.0 Safari/537.36'


<!doctype html>
<html lang="zh-TW">
<head>
	<meta http-equiv="Content-Language" content="zh-TW" />
    <meta http-equiv="Cache-Control" content="no-siteapp">
    <meta http-equiv="Cache-Control" content="no-transform">
    <meta http-equiv="Content-Type" content="text/html; charset=utf8" />
    <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=5">
    <title>從地下城到遊戲帝國101看書,從地下城到遊戲帝國無錯小說,從地下城到遊戲帝國最新章節,從地下城到遊戲帝國結局-101看書</title>
    <meta name="keywords" content="從地下城到遊戲帝國在線閱讀,從地下城到遊戲帝國無錯字無亂序精修章節,從地下城到遊戲帝國,從地下城到遊戲帝國101看書,從地下城到遊戲帝國101無彈窗廣告,從地下城到遊戲帝國完本大結局,101看書" />
    <meta name="description" content="101看書為您提供佚名創作的其他類型小說《從地下城到遊戲帝國》最新章節：新書《從末世崛起的萬界武聖》已經發布！,關鍵詞：【遊戲製作】+【領主種田】+【第四天災】。<br />
&emsp;&emsp;——<br />
&emsp;&emsp;大災變後。<br />
&emsp;&emsp;這世界有兩種特殊職業：<br />
&emsp;&emsp;「冒險家」與「地下城領主」。<br />
&emsp;&emsp;【冒險家】：永遠好奇追求刺激，如饑似渴搜尋著地下城，挑戰秘境以獲得聲望、寶藏、或純粹的快樂。<br />
&emsp;&emsp;【地下城領主】：他們陰險狡詐、野心勃勃，絞盡腦汁打造秘境、收割冒險家，從而獲取力量與財富。<br />
&emsp;&emsp;二者關係緊密。<br />
&emsp;&emsp;又充滿了博弈！<br />
&emsp;&emsp;當世界被惡意滿滿的單調迷宮充斥。<br />
&emsp;&emsp;一座畫風截然不同的地下城橫空出世，讓所有冒險家都深陷其中為之瘋狂。<br />
&emsp;&emsp;因為。<br />
&emsp;&emsp;在這裡。<br />
&emsp;&emsp;他們曾直面喪屍狂潮解開生化末日之謎，也曾在葦名城飛檐走壁與劍聖生死交鋒。<br />
&emsp;&emsp;他們曾在交界地女武神的刀鋒之下起舞，也曾在亞楠狩獵古神，挑戰超越人類認知的恐怖與神秘。<br />
&emsp;&emsp;他們曾成為鐵馭遨遊雲霄，體驗極致的機甲戰鬥，也曾重走西遊大鬧天宮，與法天象地的二郎真神巔峰鬥法。<br />
&emsp;&emsp;而我。<br />
&emsp;&emsp;正是這座傳奇地下城的締造者，奇蹟城的領主——齊霽！" />
	
	<meta property="og:type" content="novel">
    <meta property="og:book_id" content="12544">
    <meta property="og:title" content="從地下城到遊戲帝國">
    <meta property="og:image" content="https://101kks.com/files/article/image/12/12544/12544s.jpg">
    <meta property="og:description" content="關鍵詞：【遊戲製作】+【領主種田】+【第四天災】。<br />
&emsp;&emsp;——<br />
&emsp;&emsp;大災變後。<br />
&emsp;&emsp;這世界有兩種特殊職業：<br />
&emsp;&emsp;「冒險家」與「地下城領主」。<br />
&emsp;&emsp;【冒險家】：永遠好奇追求刺激，如饑似渴搜尋著地下城，挑戰秘境以獲得聲望、寶藏、或純粹的快樂。<br />
&emsp;&emsp;【地下城領主】：他們陰險狡詐、野心勃勃，絞盡腦汁打造秘境、收割冒險家，從而獲取力量與財富。<br />
&emsp;&emsp;二者關係緊密。<br />
&emsp;&emsp;又充滿了博弈！<br />
&emsp;&emsp;當世界被惡意滿滿的單調迷宮充斥。<br />
&emsp;&emsp;一座畫風截然不同的地下城橫空出世，讓所有冒險家都深陷其中為之瘋狂。<br />
&emsp;&emsp;因為。<br />
&emsp;&emsp;在這裡。<br />
&emsp;&emsp;他們曾直面喪屍狂潮解開生化末日之謎，也曾在葦名城飛檐走壁與劍聖生死交鋒。<br />
&emsp;&emsp;他們曾在交界地女武神的刀鋒之下起舞，也曾在亞楠狩獵古神，挑戰超越人類認知的恐怖與神秘。<br />
&emsp;&emsp;他們曾成為鐵馭遨遊雲霄，體驗極致的機甲戰鬥，也曾重走西遊大鬧天宮，與法天象地的二郎真神巔峰鬥法。<br />
&emsp;&emsp;而我。<br />
&emsp;&emsp;正是這座傳奇地下城的締造者，奇蹟城的領主——齊霽！">
    <meta property="og:url" content="https://101kks.com/book/12544.html">
    <meta property="og:novel:category" content="其他類型">
    <meta property="og:novel:author" content="佚名">
    <meta property="og:novel:book_name" content="從地下城到遊戲帝國">
    <meta property="og:novel:read_url" content="https://101kks.com/book/12544/index.html">
    <meta property="og:novel:latest_chapter_name" content="新書《從末世崛起的萬界武聖》已經發布！">
    <meta property="og:novel:latest_chapter_url" content="https://101kks.com/txt/12544/11210767.html">
    <meta property="og:novel:update_time" content="2025-12-31 12:28:20">
    <meta property="og:novel:status" content="全本">	
	<meta name="google-adsense-account" content="ca-pub-4729310984142201">
    <link rel="stylesheet" type="text/css" href="/css/style.css">
    <link rel="stylesheet" type="text/css" href="/css/comments.css">
      <link rel="stylesheet" type="text/css" href="/css/iconfont/iconfont.css">
    <link rel="shortcut icon" href="/favicon.ico"/>
<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=ca-pub-4729310984142201" crossorigin="anonymous"></script>
    <script>
        var bookinfo = {
            pageType: 1,
            pageVer: '20230927',
            articleid: '12544',
            articlename: '從地下城到遊戲帝國',
            siteName: '101看書',
            site: 'https://101kks.com',
            sortName: '其他類型',
            sortUrl: '/novels/class/10_1.html',
            author: '佚名',
            tags: '遊戲異界,第四天災,至尊流,思路清奇,輕鬆,'
        };
        
        // 投票功能
        function do_vote(articleid) {
            var vote_result = $('#vote_result');
            var btn_vote = $('#btn_vote');
            
            // 禁用按鈕防止重複點擊
            btn_vote.attr('disabled', true).text('投票中...');
            vote_result.text('');
            
            // 檢查用戶是否登錄
            if (typeof jieqiUserId === 'undefined' || jieqiUserId == 0) {
                vote_result.text('請先登錄後再投票');
                btn_vote.attr('disabled', false).text('投推薦票');
                return;
            }
            
            // AJAX投票
            $.ajax({
                type: 'POST',
                url: '/modules/article/uservote.php',
                data: {
                    id: articleid,
                    num: 1 // 默認投票1張
                },
                dataType: 'html',
                success: function(data) {
                    // 處理返回結果
                    if (data.indexOf('投票成功') !== -1 || data.indexOf('LANG_DO_SUCCESS') !== -1) {
                        vote_result.text('投票成功！感謝您的支持');
                        // 更新頁面上的投票數顯示
                        var vote_count_elem = document.querySelector('.bookimg2 div span:last-child');
                        if (vote_count_elem) {
                            var current_votes = parseInt(vote_count_elem.textContent);
                            if (!isNaN(current_votes)) {
                                vote_count_elem.textContent = current_votes + 1;
                            }
                        }
                    } else {
                        // 提取錯誤信息
                        var error_msg = data.match(/<p[^>]*>(.*?)<\/p>/i);
                        if (error_msg && error_msg[1]) {
                            vote_result.text(error_msg[1]);
                        } else {
                            vote_result.text('投票失敗，請稍後重試');
                        }
                    }
                },
                error: function() {
                    vote_result.text('網絡錯誤，請稍後重試');
                },
                complete: function() {
                    // 恢復按鈕
                    btn_vote.attr('disabled', false).text('投推薦票');
                }
            });
        }
    </script>
</head>
<body>
<div id="pageheadermenu">
    <header></header>
</div>
<div class="container">
    <ul class="row">
        <li class="col-8">
            <div class="mybox">
                <h3 class="mytitle">
                    <div class="bread">
                        <a href="https://101kks.com">首頁</a> &gt; <a
                            href="/novels/class/10_1.html">其他類型</a> &gt; <a
                            href="https://101kks.com/book/12544.html">從地下城到遊戲帝國</a>
                    </div>
                </h3>
                <div class="bookbox">
                    <div class="bookimg2">
                        <span class="status1"></span>
                        <img src="https://101kks.com/files/article/image/12/12544/12544s.jpg" title="從地下城到遊戲帝國" alt="從地下城到遊戲帝國">
                        <div style="position: absolute; bottom: 10px; left: 10px; right: 10px; background: rgba(0, 0, 0, 0.7); color: white; padding: 6px 12px; border-radius: 15px; font-size: 12px; font-weight: bold; text-align: center;">
                            <span style="margin-right: 4px;">👍</span>
                            推薦票：<span style="font-size: 14px; color: #fff9c4; font-weight: 800;">0</span>
                        </div>
                    </div>
                    
                    <div class="booknav2">
                        <h1><a href="https://101kks.com/book/12544.html">從地下城到遊戲帝國</a></h1>                  
                       <p>作者：<a href="https://101kks.com/author/佚名.html" title="佚名">佚名</a></p>
                        <p>分類：<a href="/novels/class/10_1.html" title="其他類型">其他類型</a></p>
                        <p>212.28萬字 | 全本</p>
                       
                        <p>更新：2025-12-31 <a href="/newmessage.php?tosys=1&title=未更新&content=https://101kks.com/txt/12544/11210767.html"  class="btn-urge">↔催更</a></p>
                        <div class="sharebtn"><div class="line-it-button" data-lang="zh_Hant" data-type="share-a" data-env="REAL" data-url="https://101kks.com/book/12544.html" data-color="default" data-size="small" data-count="false" data-ver="3" style="display: none;"></div>
                        <script src="https://www.line-website.com/social-plugins/js/thirdparty/loader.min.js" async="async" defer="defer"></script>
                        <iframe src="https://www.facebook.com/plugins/share_button.php?href=https%3A%2F%2F69shux.com%2Fbook%2F12544.html&amp;width=138&amp;layout=button&amp;action=like&amp;size=small&amp;share=true&amp;height=65&amp;appId=" width="168" height="28" style="border:none;overflow:hidden" scrolling="no" frameborder="0" allowfullscreen="true" allow="autoplay; clipboard-write; encrypted-media; picture-in-picture; web-share"></iframe>
                       </div>
                     </div>

                    <div class="addbtn">
                            <a class="btn" href="https://101kks.com/book/12544/index.html">開始閱讀</a>
                            <a class="btn" id="a_addbookcase" href="javascript:;" onclick="addbookcase_info();">加入書架</a>
                            <a id="bookcase" href="#" class='btn white' style="display:none">進入書籤</a>
                            <a class="btn" id="btn_vote" href="javascript:;" onclick="do_vote(12544);">投推薦票</a>
                        </div>
                        <div id="vote_result" style="clear: both; padding: 10px 10px 0 10px; color: #f00; font-size: 14px;"></div>
                </div>
            </div>

            <div class="mybox">
                <div class="infotag  clearfix" id="infotag">
                    
                    <h3 class="tagtitle">標籤</h3>
                    <ul class="tagul" id="tagul">
                    </ul>   
                </div>
                <ul class="tabs clearfix">
                    <li class="active"><a id="li_chapters" href="javascript:;">
                            <i class="iconfont icon-list"></i>
                            目錄</a>
                    </li>
                    <li><a id="li_info" href="javascript:;">
                        <i class="iconfont icon-Info"></i>
                        簡介</a>
                    </li>
                    <li><a id="li_reviews" href="javascript:;">
                        <i class="iconfont icon-chat"></i>
                        書評</a>
                    </li>
                </ul>
                <div class="tabsnav">
                    <div id="tab_chapters" >
                      <ul class="qustime">
	
	 <li><a href="https://101kks.com/txt/12544/11210767.html"  ><span>新書《從末世崛起的萬界武聖》已經發布！</span><small>2025-12-31</small></a></li><li>
	
	 <li><a href="https://101kks.com/txt/12544/8898221.html"  ><span>完本感言</span><small>2025-09-24</small></a></li><li>
	
	 <li><a href="https://101kks.com/txt/12544/8898219.html"  ><span>第316章 奇蹟大帝！奇蹟帝國！（大結局）</span><small>2025-09-24</small></a></li><li>
	
	 <li><a href="https://101kks.com/txt/12544/8898217.html"  ><span>第315章 碾壓黑暗大帝！突破七階！</span><small>2025-09-24</small></a></li><li>
	
	 <li><a href="https://101kks.com/txt/12544/8898216.html"  ><span>第314章 底牌盡出！各顯神通！</span><small>2025-09-24</small></a></li><li>
	
	 <li><a href="https://101kks.com/txt/12544/8898214.html"  ><span>第313章 終極戰役</span><small>2025-09-24</small></a></li><li>
	
</ul>
                    </div>
                    <div id="tab_info" style="display: none;">
                        <ul class="infolist">
                            <li>212.28萬字<span>字數</span></li>
                            <li>323<span>章節數</span></li>
                        </ul>
                        <div class="navtxt">
                          <p>關鍵詞：【遊戲製作】+【領主種田】+【第四天災】。<br />
&emsp;&emsp;——<br />
&emsp;&emsp;大災變後。<br />
&emsp;&emsp;這世界有兩種特殊職業：<br />
&emsp;&emsp;「冒險家」與「地下城領主」。<br />
&emsp;&emsp;【冒險家】：永遠好奇追求刺激，如饑似渴搜尋著地下城，挑戰秘境以獲得聲望、寶藏、或純粹的快樂。<br />
&emsp;&emsp;【地下城領主】：他們陰險狡詐、野心勃勃，絞盡腦汁打造秘境、收割冒險家，從而獲取力量與財富。<br />
&emsp;&emsp;二者關係緊密。<br />
&emsp;&emsp;又充滿了博弈！<br />
&emsp;&emsp;當世界被惡意滿滿的單調迷宮充斥。<br />
&emsp;&emsp;一座畫風截然不同的地下城橫空出世，讓所有冒險家都深陷其中為之瘋狂。<br />
&emsp;&emsp;因為。<br />
&emsp;&emsp;在這裡。<br />
&emsp;&emsp;他們曾直面喪屍狂潮解開生化末日之謎，也曾在葦名城飛檐走壁與劍聖生死交鋒。<br />
&emsp;&emsp;他們曾在交界地女武神的刀鋒之下起舞，也曾在亞楠狩獵古神，挑戰超越人類認知的恐怖與神秘。<br />
&emsp;&emsp;他們曾成為鐵馭遨遊雲霄，體驗極致的機甲戰鬥，也曾重走西遊大鬧天宮，與法天象地的二郎真神巔峰鬥法。<br />
&emsp;&emsp;而我。<br />
&emsp;&emsp;正是這座傳奇地下城的締造者，奇蹟城的領主——齊霽！</p>
                          小說關鍵詞：從地下城到遊戲帝國無彈窗,從地下城到遊戲帝國無亂序,從地下城到遊戲帝國小說,從地下城到遊戲帝國佚名,從地下城到遊戲帝國101,從地下城到遊戲帝國月下藏鋒,從地下城到遊戲帝國最新章節閱讀
                        </div>
                    </div>
                    <div id="tab_reviews" style="display: none;">
                            <section class="review-panel ">
                                <div class="review-list fullw">
                            
                            
<div class="swiper-slide">
<div class="review-item">
<div class="review-info">
<div class="user">
 <img src="https://101kks.com/files/system/avatar/0/3s.jpg" onerror="this.src='https://ui-avatars.com/api/?name=%E6%BD%AE%E6%B1%90'" alt="潮汐">
<strong>潮汐</strong>
</div>
<div class="c_tag">
    <span class="c_label">時間：</span><span class="c_value">2026-01-04 21:18:42</span>
    <span class="c_label  hide720">點擊：</span><span class="c_value  hide720">10</span>
    <span class="c_label  hide720">回覆：</span><span class="c_value  hide720">0</span>
</div>
</div>
<p class="review-text">
.　　跟風《這陰間地下城誰設計的》，地下城經營類型變種遊戲文抄文。<br />
　　繼《這陰間地下城誰設計的》之後，又一本主角穿越成地下城城主，然後文抄遊戲製作地下城秘境，以供冒險者（玩家）交互遊玩的小說。<br />
　　這種世界觀設定，確實非常方便在奇幻異界以遊戲為載體進行文化入侵。<br />
　　而且感覺受眾群，或者說基礎盤，還比較穩定。
</p>
<div class="c_tag">
    <span class="c_label fr"><a href="/reviews/show/174.html#frmpost" >[我要回復]</a></span><span class="c_value"></span>
</div>
</div>
</div>

                            
								</div>
                            </section>
                            <!-- <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/Trumbowyg/2.21.0/ui/trumbowyg.min.css"> -->
                            <!-- <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/Trumbowyg/2.21.0/plugins/emoji/ui/trumbowyg.emoji.min.css"> -->
                            <section class="comment-list" data-bookid="12544">
                                <div class="section-header">
                                    <h3>發表書評</h3>
                                </div>
                            <span>喜歡《從地下城到遊戲帝國》的讀者可以在此發表評論!不但可以與其他書友分享樂趣，也可以增加積分和經驗！</span>

                            <div id="lnwcomeditor"><div class="comment-area">
                                <div class="ratemain">
                                <script type="text/javascript" src="https://101kks.com/scripts/rating.js"></script>
                                <div class="ratediv"><b class="fl">評分：</b>
                                    <div class="rateblock" id="rate_star">
                                    <script type="text/javascript">
                                      showRating(10, 8.0, 'rating', '12544');
                                      function rating(score, id){
                                        Ajax.Tip('/modules/article/rating.php?score='+score+'&id='+id, {method: 'POST', eid: 'rate_star'});
                                      }
                                    </script>
                                    </div>
                                    <span class="ratenum">8.0</span> <span class="gray">(1人已評)</span>
                                </div>
                                </div>
                                <form name="frmreview" id="frmreview" method="post" action="/reviews/12544.html">
                                    <input type="hidden" class="txt_block" name="ptitle" id="ptitle" size="60" maxlength="60" value="" placeholder="標題"/>
                                        <textarea id="comments" class="txt_block"  name="pcontent"  minlength="10" maxlength="200" placeholder="友善評論，文明發言。大家可以表達看法，意見，但不要有侮辱，謾罵類的評論，即便是我們不喜歡的。"></textarea>
                                        <input type="hidden" name="act" value="newpost" />
                                        <button href="javascript:;" id="btn_sendtxt" type="button" class="button" onclick="window.location.href='/login.php'">請先登錄</button>
                            </form>
                            </div></div>
                            
                            <ul>
                            </ul>
                            </section>
                            </section>
                    </div>
                </div>
                <a class="btn more-btn" href="https://101kks.com/book/12544/index.html">完整目錄</a>
            </div>
        </li>
        <li class="col-4">
            <div class="mybox">
                <h3 class="mytitle">本周最強</h3>
                <ul class="tabs tabshot clearfix">
                    <li class="active"><a href="javascript:;"><i class="iconfont icon-hot"></i>熱門</a></li>
                    <li><a href="javascript:;"><i class="iconfont icon-hot"></i>完本</a></li>
                </ul>
                <div class="tabsnav">
                    <div class="ranking">
                    <ul>

<li class="active">
    <a href="https://101kks.com/book/30974.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>呼吸都能變強，你說我最弱天賦？</h3>
            <h4>本周最強</h4>
            <p>現代都市.記憶的海</p>
        </div>
        <div class="rank_right">
            <div class="imgbox2">
                <img src="https://101kks.com/files/article/image/30/30974/30974s.jpg" alt="呼吸都能變強，你說我最弱天賦？">
            </div>
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/23891.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>巫師：我在兩界當泰坦</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/27330.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>木葉宇智波，開局硬槓木葉！</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/24527.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>當了五十億年太陽，我修仙了</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/31179.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>人在漫威賣扭蛋，開局托尼變空我</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/31001.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>全民公路求生：我的房車無限進化</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/30788.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>天賦等級加一，超然的群星巫師</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/6667.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>假死騙我離婚，重生我火化前妻</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/116.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>御獸從零分開始</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/2391.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>離婚是吧，我轉身迎娶百億女總裁</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/29317.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>四合院開局保定打斷腿</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/17275.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>全民拾荒求生！我能看到全圖資源</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/24984.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>四合院：我有10個兒子</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/9266.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>苟在初聖魔門當人材</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/12422.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>反派：開局拿下蘿莉女主</h3>
        </div>
        <div class="rank_right">
            <span>連載</span>
        </div>
    </a>
</li>

</ul>



                    </div>
                    <div class="ranking" style="display: none;">
                   <ul>

<li class="active">
    <a href="https://101kks.com/book/2431.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>我真沒想下圍棋啊！</h3>
            <h4>本周最強</h4>
            <p>現代都市.山中土塊</p>
        </div>
        <div class="rank_right">
            <div class="imgbox2">
                <img src="https://101kks.com/files/article/image/2/2431/2431s.jpg" alt="我真沒想下圍棋啊！">
            </div>
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/10230.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>海賊：從撿到紅髮斷臂開始</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/10379.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>說好開發消消樂，地球戰爭什麼鬼</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/38.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>高武紀元</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/7994.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>超維度玩家</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/71.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>赤心巡天</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/14032.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>火影：我在暗部苟成超影</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/10128.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>天災信使</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/59.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>輪迴樂園</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/151.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>守序暴君</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/3368.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>成影帝了，系統才加載完</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/73.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>萬相之王</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/96.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>普羅之主</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/5455.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>牧神記</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

 <li>
    <a href="https://101kks.com/book/228.html">
        <div class="rank_left">
            <h3 class="ranktit ellipsis_1"><span></span>我有一個修仙世界</h3>
        </div>
        <div class="rank_right">
            <span>全本</span>
        </div>
    </a>
</li>

</ul>



                    </div>
                </div>
            </div>
        </li>
    </ul> 
</div>

<div id="pagefootermenu"></div>
<script src="/js/jquery.min.js"></script>
<script src="/js/newmenu.js?v=20240040801"></script>
<div class="modelbg"></div>
<script src="/js/chinese.js"></script>
<script src="/js/newread.js?v=20231014"></script>
<script>tongji();</script>
<script>
    function addbookcase_info()
    {
        addbookcase(12544,0);
        CheckBookCase();
    }
    function CheckBookCase(){
        $.ajax({
            type: 'GET',
            url: '/modules/article/checkbookcase.php?bid=12544',
            dataType: 'html',
            success:function(data){
               if(parseInt(data)>0)
               {
                $('#a_addbookcase').text("移出書架");
                $('#a_addbookcase').attr("onclick","act_delete("+data+");");
                }
               else{
                $('#a_addbookcase').text("加入書架");
               }
        }
        });
    }
    $(document).ready(function()
    {
        CheckBookCase();
        var url = window.location.href;  
        if(url.indexOf("#reviews") >= 0 ) { 
            document.getElementById('li_reviews').click();
        } 
		if(jieqiUserId != 0 && jieqiUserName != '' && (document.cookie.indexOf('PHPSESSID') != -1 || jieqiUserPassword != '')){  
			var btn=document.getElementById('btn_sendtxt');
			btn.innerHTML="發表書評";
			btn.type='submit';
			btn.onclick='';	
		};

		
    }
    );
</script>
<script type="text/javascript">
    function act_delete(bookid){
        if(confirm('確實要將本書移出書架麼？')) 
         {
                var token= "06e51c51c0ae6eb93e1a73307ae754b0";
                $.post({
                 contenttype:'application/x-www-form-urlencoded; charset=UTF-8',
                  data:{act:'delete',bid:bookid,jieqi_token:token},
                 url: "/modules/article/bookcase.php?bid="+bookid,
                  datatype: "html",
                 success: function(data){
                    $('#a_addbookcase').text("加入書架");
                    $('#a_addbookcase').attr("onclick","addbookcase_info();");
                    alert('刪除成功!');
                    console.log(data);
                 }
                });
        }
        return false;
    }
    </script>

<script defer src="https://static.cloudflareinsights.com/beacon.min.js/vcd15cbe7772f49c399c6a5babf22c1241717689176015" integrity="sha512-ZpsOmlRQV6y907TI0dKBHq9Md29nnaEIPlkf84rnaERnq6zvWvPUqr2ft8M1aS28oN72PdrCzSjY4U6VaAw1EQ==" data-cf-beacon='{"rayId":"9c74fc26f816ec4e","serverTiming":{"name":{"cfExtPri":true,"cfEdge":true,"cfOrigin":true,"cfL4":true,"cfSpeedBrain":true,"cfCacheStatus":true}},"version":"2025.9.1","token":"a20e3bde50584dad811938680261d00f"}' crossorigin="anonymous"></script>
</body>
</html> 

curl 'https://101kks.com/txt/12544/8898221.html' \
  -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \
  -H 'accept-language: ru,en;q=0.9' \
  -b 'zh_choose=t; cf_clearance=HCBZS3y57ehal5qQazG39l3Upvd1WekwKBJ0fwtm8k0-1769985973-1.2.1.1-.N8VCuX9TquQoNHyfnedDv.1KgONZM1dEQXc8pWRK9CgW19anSXcVzeKl.mdRJVaSfydMiHDB8wG4oBexjblPt.8WbLvF6BJFxD0ZPcVSYQd946.QYHSiRzoWcbQ0NStk355oSQVoeeBlRHX7llBeyOR1tcxe8d2xXiBgn1ZckpLmPzdi8S0zPCwr.kGsjGaZQM5jv98.X2W07eiG6_bK6Hulr3hqTavFgy3IVEjHFo; _ga=GA1.1.871789781.1769985973; _ga_DMRW3FNJ29=GS2.1.s1769985973$o1$g1$t1769986218$j28$l0$h0' \
  -H 'priority: u=0, i' \
  -H 'referer: https://101kks.com/book/12544.html' \
  -H 'sec-ch-ua: "Chromium";v="136", "YaBrowser";v="25.6", "Not.A/Brand";v="99", "Yowser";v="2.5"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: document' \
  -H 'sec-fetch-mode: navigate' \
  -H 'sec-fetch-site: same-origin' \
  -H 'sec-fetch-user: ?1' \
  -H 'upgrade-insecure-requests: 1' \
  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 YaBrowser/25.6.0.0 Safari/537.36' 

  <!DOCTYPE html>
<html lang="zh-TW">
  
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf8" />
	<meta http-equiv="Content-Language" content="zh-TW" />
    <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=5">
    <title>完本感言-佚名-其他類型-101看書</title>
    <meta name="keywords" content="完本感言,從地下城到遊戲帝國,從地下城到遊戲帝國101看書,從地下城到遊戲帝國101kan,從地下城到遊戲帝國無亂序" />
    <meta name="description" content="完本感言-從地下城到遊戲帝國101看書-最新章節無錯/無彈窗廣告閱讀/無亂序-101看書" />
	<meta name="google-adsense-account" content="ca-pub-4729310984142201">
<script src="https://cdnjs.cloudflare.com/ajax/libs/crypto-js/3.1.9-1/crypto-js.js"></script>

    <link rel="prefetch" href="https://101kks.com/txt/12544/11210767.html" />
    <link rel="stylesheet" type="text/css" href="/css/yuedu.css">
    <link rel="shortcut icon" href="/favicon.ico" />
	<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=ca-pub-4729310984142201" crossorigin="anonymous"></script>

    <script src="/js/otherad.js"></script>
   <script>loadAdv(1000, 0);</script>
   <script>var bookinfo = {
        pageType: 3,
        pageVer: '202308002',
        siteName: '101看書',
        site: 'https://101kks.com',
        articleid: '12544',
        chapterid: '8898221',
        articlename: '',
        chaptername: '完本感言',
        index_page: 'https://101kks.com/book/12544/index.html',
        sortName: '其他類型',
        sortUrl: '/novels/class/10_1.html',
        author: '佚名',
        preview_page: "https://101kks.com/txt/12544/8898219.html",
        next_page: "https://101kks.com/txt/12544/11210767.html"
      };</script>
  <script>if (window.top != window.self) {window.top.location = window.self.location;}</script>
  <script>
  eval(function(p,a,c,k,e,r){e=function(c){return c.toString(a)};if(!''.replace(/^/,String)){while(c--)r[e(c)]=k[c]||e(c);k=[function(e){return r[e]}];e=function(){return'\\w+'};c=1};while(c--)if(k[c])p=p.replace(new RegExp('\\b'+e(c)+'\\b','g'),k[c]);return p}('9 4(a,b,c,d){5.e("f"+d).g.h=\'i\';b=1.2.3.6(b.7(8,0));c=1.2.3.6(c.7(8,0));a=1.j.4(a,b,{k:c,l:1.m.n,}).o(1.2.3);5.p(a)}',26,26,'|CryptoJS|enc|Utf8|decrypt|document|parse|padStart|16|function|||||getElementById|tips_|style|display|none|AES|iv|padding|pad|Pkcs7|toString|write'.split('|'),0,{}))
  </script>

  </head>
  
  <body>
    <div id="pageheadermenu">
      <header></header>
    </div>
    <div class="container" id="container">
      <div class="mybox">
	   <!-- <div class="yueduad1"> -->
            <!-- <script>loadAdv(1, 1);</script> -->
        <!-- </div> -->
        <h3 class="mytitle hide720">
          <div class="bread">
            <a href="https://101kks.com">首頁</a>>
            <a href="/novels/class/10_1.html">其他類型</a>>
            <a href="https://101kks.com/book/12544/index.html">目錄頁</a>> 完本感言</div></h3>
        <div class="tools">
          <ul>
            <li>
              <a href="https://101kks.com/book/12544.html">
                <span>書頁</span></a>
            </li>
            <li>
              <span>
                <a id="a_addbookcase" href="javascript:;" onclick="addbookcase(12544,8898221);">收藏</a></span>
            </li>
            <li>
              <a href="https://101kks.com/book/12544/index.html">
                <i class="iconfont icon-mulu"></i>
                <span>目錄</span></a>
            </li>
            <li>
              <a href="javascript:;" class="setboxbtn">
                <i class="iconfont icon-setting"></i>
                <span>設置</span></a>
            </li>
            <li>
              <a href="javascript:;" onclick="setbg(setbg)">
                <i class="iconfont icon-moon"></i>
                <span class="bgtxt">黑夜</span></a>
            </li>
          </ul>
        </div>

        <div class="txtnav">
          <h1>完本感言</h1>
          <div class="txtinfo hide720">
            <span>2025-09-24 19:52:53</span>
            <span>作者： 佚名</span></div>
			<div class="chase-book-btn">
              <button class="readButton" id="start-readButton-link" data-umami-event="藍雨" data-umami-action="點擊" data-spk="lanyuf">藍雨</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="麻豆" data-umami-action="點擊" data-spk="madoufp_wenrou">麻豆</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="小玲" data-umami-action="點擊" data-spk="zhilingf">小玲</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="安寧" data-umami-action="點擊" data-spk="aningfp">安寧</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="零八" data-umami-action="點擊" data-spk="linbaf_qingxin">零八</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="小妖" data-umami-action="點擊" data-spk="xiyaof_qingxin">小妖</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="考拉" data-umami-action="點擊" data-spk="kaolam_diantai">考拉</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="小軍" data-umami-action="點擊" data-spk="xijunma">小軍</button>
              <button class="readButton" id="start-readButton-link" data-umami-event="秋木" data-umami-action="點擊" data-spk="qiumum_0gushi">秋木</button>
              </div>
			  
           
            <div class="txtcenter">
			<script>loadAdv(2, 0);</script>
			</div>
          <div class="txtad">
          </div>
		   <div id="txtcontent">
		   &emsp;&emsp;感謝大家。<br />
<br />
&emsp;&emsp;本書寫到這裡就完本了。<br />
<br />
&emsp;&emsp;一百六十萬字不到，這肯定與最初設想的預期篇幅有較大差距，但已經是在現有條件下完成度相對較高的結局了。<br />
<br />
&emsp;&emsp;書的遺憾與不足是有的。<br />
<br />
&emsp;&emsp;我認為網文作品風格大體能分三類，最上乘的寫法是靠故事和人物撐起來的，其次是靠設定的趣味來吸引讀者，最後則是依靠情緒比如贅婿歪嘴龍王什麼的。<br />
<br />
&emsp;&emsp;當然這只是個人的一點淺見。<br />
<br />
&emsp;&emsp;三者本身也並不互相排斥，有些書確實能做到既有好故事、又有趣味性、還很能拉情緒，但這樣的書恐怕鳳毛麟角並不好寫。<br />
<br />
&emsp;&emsp;按照這種分類進行劃分。<br />
<br />
&emsp;&emsp;本書毫無疑問是靠設定的趣味性來吸引讀者的。<br />
<br />
&emsp;&emsp;只是這種書以及所有偏腦洞的寫法，前中期比較容易出成績，但中後期想要繼續維持新鮮感和趣味性難度卻是很大。<br />
<br />
&emsp;&emsp;又因為作者的個人原因。<br />
<br />
&emsp;&emsp;比如過於倉促的開書，上本寫了近三百五十萬字，沒休息一天直接無縫銜接就開了本書。<br />
<br />
&emsp;&emsp;外加上架爆更期間剛好與備婚和結婚重迭，導致我在設定上的打磨程度不太高，所以導致後期展開遇到困難。<br />
<br />
&emsp;&emsp;此外還有一個比較致命的點就是，從新書期開始本書讀者就是更喜歡看秘境內容，這導致作者過份聚焦遊戲劇情，沒敢分出筆墨去鋪墊世界觀和中長線故事，因為每次一寫秘境外的內容數據就嘩嘩掉。<br />
<br />
&emsp;&emsp;最後的結果就是越不寫或簡化外部世界、領地開拓過程，這部分劇情就越沒吸引力，越沒吸引力就越不敢輕易分出筆墨鋪墊豐滿，當中後期遊戲新鮮感褪去，就缺乏中長線吸引人看下去的期待感。<br />
<br />
&emsp;&emsp;這也是本書沒有寫到預期篇幅的最主要原因。<br />
<br />
&emsp;&emsp;當然。<br />
<br />
&emsp;&emsp;題材的問題並非不能解決。<br />
<br />
&emsp;&emsp;我覺得歸根到底最主要的原因還是準備不足、時間不夠，寫作過程中太多私人事情消耗了一部分精力導致思考時間太少。<br />
<br />
&emsp;&emsp;個人覺得在這本書上犯的最大錯誤太過於倉促的創建本書，以後一定要吸取教訓，不做好準備、不擬好細綱絕不開書。<br />
<br />
&emsp;&emsp;回到這本書。<br /><br />  <br />
<br />
&emsp;&emsp;我之所以會寫這本書。<br />
<br />
&emsp;&emsp;其實最初的靈感來自《系統的黑科技網吧》，相信大家也能看出來，本書的秘境大廳設定就是早些年網吧文的網吧，我個人覺得這種設定是很有趣的。<br />
<br />
&emsp;&emsp;不過我把背景設定在一個偏西幻的世界似乎略顯小眾，我有時候也會想，假如索性寫一個純遊戲製作文，或者簡化一點不融入這麼多設定，受眾成績或許會更好一點。<br />
<br />
&emsp;&emsp;這本書正文到此就告一段落。<br />
<br />
&emsp;&emsp;未來如果有非常感興趣的遊戲想分享、或許會以番外方式進行呈現。<br />
<br />
&emsp;&emsp;至於新書。<br />
<br />
&emsp;&emsp;這次肯定沒有這麼快。<br />
<br />
&emsp;&emsp;作者菌不想重蹈覆轍再不無縫銜接開書了。<br />
<br />
&emsp;&emsp;預計會休整一個月左右，大約在十月的月底才會開出來，而在此期間我會努力思考、復盤、看書，更重要的是做好選題寫大綱。<br />
<br />
&emsp;&emsp;下本書的目標是既要有較為有趣的設定、更要有足夠紮實的世界觀、人物以及故事。 <div class="txtad"><script>loadAdv(10,0);</script></div><br />
<br />
&emsp;&emsp;感謝每一位喜歡本書的朋友。<br />
<br />
&emsp;&emsp;作者某信是xxybz1234<br />
<br />
&emsp;&emsp;如果對新書進度感興趣，或者想要加群的朋友，都可以加我進行諮詢哈。<br />
<br />
&emsp;&emsp;最後再一次感謝每一位打賞、訂閱過本書、支持過藏鋒老弟的朋友，你們每一位都是藏鋒老弟的衣食父母，我沒能以最好的狀態寫完本書實屬遺憾和抱歉。<br />
<br />
&emsp;&emsp;寫書是我的職業也是唯一擅長的事。<br />
<br />
&emsp;&emsp;我從大一開始寫書到現在十多年，巔峰期在外站也曾寫過幾本成績非常不錯，至少遠超本書各方面數據數倍的書，但這兩年似乎落入了相對的低谷狀態欠佳。<br />
<br />
&emsp;&emsp;不過雖說如此。<br />
<br />
&emsp;&emsp;我會努力改進自己。<br />
<br />
&emsp;&emsp;因為我仍沒有放棄寫出一本好書的理想或者書執念，希望在這個市場變化極快的時代，找到更適合自己的題材與故事。<br />
<br />
&emsp;&emsp;有緣的話。<br />
<br />
&emsp;&emsp;下一本書再見。<br />
<br />
&emsp;&emsp;最後再一次感謝大家的支持與理解！(本章完)
		   </div>
		  <br />
         <div class="txtcenter">
         <script>loadAdv(3, 0);</script>
         </div>
        </div>
        <div class="page1">
          <a href="https://101kks.com/txt/12544/8898219.html">上一章</a>
          <a id="a_addbookcase" href="javascript:;" onclick="addbookcase(12544,8898221);">書籤</a>
          <a href="https://101kks.com/book/12544/index.html" title="從地下城到遊戲帝國最新章節列表">目錄</a>
          <a href="https://101kks.com/txt/12544/11210767.html">下一章</a></div>
        <br />
        <div class="txtcenter">
         <script>loadAdv(5, 0);</script>
       </div>
        <div id="baocuo" class="baocuo"></div>
      </div>
	  <div id="tuijian" class="tuijian"></div>
      <div class="yuedutuijian light " style="background-color:#fff"><ul>

<li><a href="https://101kks.com/book/30974.html" style="color: #222">呼吸都能變強，你說我最弱天賦？</a><span>連載</span></li>

<li><a href="https://101kks.com/book/23891.html" style="color: #222">巫師：我在兩界當泰坦</a><span>連載</span></li>

<li><a href="https://101kks.com/book/27330.html" style="color: #222">木葉宇智波，開局硬槓木葉！</a><span>連載</span></li>

<li><a href="https://101kks.com/book/24527.html" style="color: #222">當了五十億年太陽，我修仙了</a><span>連載</span></li>

<li><a href="https://101kks.com/book/31179.html" style="color: #222">人在漫威賣扭蛋，開局托尼變空我</a><span>連載</span></li>

<li><a href="https://101kks.com/book/31001.html" style="color: #222">全民公路求生：我的房車無限進化</a><span>連載</span></li>

<li><a href="https://101kks.com/book/30788.html" style="color: #222">天賦等級加一，超然的群星巫師</a><span>連載</span></li>

<li><a href="https://101kks.com/book/6667.html" style="color: #222">假死騙我離婚，重生我火化前妻</a><span>連載</span></li>

<li><a href="https://101kks.com/book/116.html" style="color: #222">御獸從零分開始</a><span>連載</span></li>

<li><a href="https://101kks.com/book/2391.html" style="color: #222">離婚是吧，我轉身迎娶百億女總裁</a><span>連載</span></li>

<li><a href="https://101kks.com/book/29317.html" style="color: #222">四合院開局保定打斷腿</a><span>連載</span></li>

<li><a href="https://101kks.com/book/17275.html" style="color: #222">全民拾荒求生！我能看到全圖資源</a><span>連載</span></li>

<li><a href="https://101kks.com/book/24984.html" style="color: #222">四合院：我有10個兒子</a><span>連載</span></li>

<li><a href="https://101kks.com/book/9266.html" style="color: #222">苟在初聖魔門當人材</a><span>連載</span></li>

<li><a href="https://101kks.com/book/12422.html" style="color: #222">反派：開局拿下蘿莉女主</a><span>連載</span></li>

</ul>


        </div></div>
    <div class="setbox">
      <div class="setli">
        <a href="javascript:;" class="setclose">關閉</a>
        <ul>
          <li>
            <label>背景</label>
            <div class="setbg">
              <a href="javascript:;" onclick="setnavbg(0)"></a>
              <a href="javascript:;" onclick="setnavbg(1)"></a>
              <a href="javascript:;" onclick="setnavbg(2)"></a>
              <a href="javascript:;" onclick="setnavbg(3)"></a>
              <a href="javascript:;" onclick="setnavbg(4)"></a>
            </div>
          </li>
          <li>
            <label>字體</label>
            <div class="setfontf">
              <a href="javascript:;" onclick="setfont(0)">雅黑</a>
              <a href="javascript:;" onclick="setfont(1)">粉圓</a>
              <a href="javascript:;" onclick="setfont(2)">手寫</a>
              <a href="javascript:;" onclick="setfont(3)">鋼筆</a></div>
          </li>
          <li>
            <label>字號</label>
            <div class="setfontsize">
              <input class="sizenum" type="text" value="22" readonly>
              <a class="cut" href="javascript:;" onclick="fontcut()">-</a>
              <a class="add" href="javascript:;" onclick="fontadd()">+</a></div>
          </li>
		  	 <li>
            <label>語速</label>
            <div class="setfontsize">
              <input class="speednum" type="text" value="100" readonly>
              <a class="cut" href="javascript:;" onclick="speechcut()">-</a>
              <a class="add" href="javascript:;" onclick="speechadd()">+</a></div>
          </li>
        </ul>
      </div>
    </div>
    <div id="goTopBtn">
      <span class="glyphicon glyphicon-chevron-up" title="返回頂部">Δ</span></div>
    <div id="pagefootermenu"></div>
    <div class="modelbg"></div>
    <script src="/js/jquery.min.js"></script>
    <script src="/js/newmenu.js"></script>
    <script src="/js/chinese.js"></script>
    <script src="/js/newread.js"></script>
    <script>Tools();</script>
	<script>tongji();</script>
    <script>var lastread = new LastRead;
      lastread.set("12544", "8898221", "從地下城到遊戲帝國", " 完本感言", "佚名", "其他類型", "https://101kks.com/files/article/image/12/12544/12544s.jpg", "全本")</script>
    <script>loadAdv(4, 0);</script><script defer src="https://static.cloudflareinsights.com/beacon.min.js/vcd15cbe7772f49c399c6a5babf22c1241717689176015" integrity="sha512-ZpsOmlRQV6y907TI0dKBHq9Md29nnaEIPlkf84rnaERnq6zvWvPUqr2ft8M1aS28oN72PdrCzSjY4U6VaAw1EQ==" data-cf-beacon='{"rayId":"9c7501d779a6e91b","serverTiming":{"name":{"cfExtPri":true,"cfEdge":true,"cfOrigin":true,"cfL4":true,"cfSpeedBrain":true,"cfCacheStatus":true}},"version":"2025.9.1","token":"a20e3bde50584dad811938680261d00f"}' crossorigin="anonymous"></script>
</body>
	

</html>