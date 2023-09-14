package utils

import (
	"regexp"
)

var (
	IpWordingPattern = regexp.MustCompile(`(?s)(ip_wording\s+=.*?;)`)
)

var (
	BizPattern           = regexp.MustCompile(`(?s)(var\s+biz\s+=.*?;)`)
	SnPattern            = regexp.MustCompile(`(?s)(var\s+sn\s+=.*?;)`)
	WxidPattern          = regexp.MustCompile(`(?s)(var\s+user_name\s+=.*?;)`)
	AppMsgIdPattern      = regexp.MustCompile(`(?s)(var\s+appmsgid\s+=.*?;)`)
	MidPattern           = regexp.MustCompile(`(?s)(var\s+mid\s+=.*?;)`)
	ReqIdPattern         = regexp.MustCompile(`(?s)(var\s+req_id\s+=.*?;)`)
	AlbumIdPattern       = regexp.MustCompile(`(?s)(album_id\s*=\s*['"].*?['"])`)
	AuthorPattern        = regexp.MustCompile(`(?s)(var\s+author\s+=.*?;)`)
	AuthorIdPattern      = regexp.MustCompile(`(?s)(var\s+author_id\s+=.*?;)`)
	SessionIdPattern     = regexp.MustCompile(`(?s)(var\s+sessionid\s+=.*?;)`)
	ArticleTypePattern   = regexp.MustCompile(`(?s)(var\s+_ori_article_type\s+=.*?;)`)
	CommentIdPattern     = regexp.MustCompile(`(?s)(var\s+comment_id\s+=.*?;)`)
	CreateTimePattern    = regexp.MustCompile(`(?s)(var\s+createTime\s+=.*?;)`)
	CopyrightStatPattern = regexp.MustCompile(`(?s)(var\s+_copyright_stat\s+=.*?;)`)
	VoiceListPattern     = regexp.MustCompile(`(?s)(voiceList\s*=\s*.*?).*?var`)
	PublicTagInfoPattern = regexp.MustCompile(`(?s)(var\s+publicTagInfo\s+=.*?];)`)

	/*-------------------------------------------------------------------*/

	CgiDataPattern     = regexp.MustCompile(`(?s)(cgiData\s+=.*?;)`) // 通过作者获取文章
	AppMagTokenPattern = regexp.MustCompile(`(?s)(appmsg_token\s+=.*?;)`)

	/*-------------------------------------------------------------------*/

	LogoUrlPattern = regexp.MustCompile(`(?s)(var\s+hd_head_img\s+=.*?;)`)
)

var (
	BizImgPattern       = regexp.MustCompile(`(?s)window.(biz\s+=\s+.*?;)`)
	SnImgPattern        = regexp.MustCompile(`(?s)window.(sn\s+=\s+.*?;)`)
	WxidImgPattern      = regexp.MustCompile(`(?s)user_name\s+=\s+.*?;`) // ---》注意
	AppMsgIdImgPattern  = regexp.MustCompile(`(?s)window.(appmsgid\s+=\s+.*?;)`)
	MidImgPattern       = regexp.MustCompile(`(?s)window.(mid\s+=\s+.*?;)`)
	ReqIdImgPattern     = regexp.MustCompile(`(?s)req_id\s+=\s+.*?;`) // ---》注意
	AlbumIdImgPattern   = regexp.MustCompile(`(?s)(album_id:.*?),`)
	AuthorImgPattern    = regexp.MustCompile(`(?s)nick_name\s+=\s+.*?;`)
	AuthorIdImgPattern  = regexp.MustCompile(`(?s)author_id\s+=\s+.*?;`)
	SessionIdImgPattern = regexp.MustCompile(`(?s)window.(sessionid\s+=\s+.*?;)`)
	//ArticleTypePattern 无
	CommentIdImgPattern  = regexp.MustCompile(`(?s)comment_id\s+=\s+.*?;`)
	CreateTimeImgPattern = regexp.MustCompile(`(?s)(var\s+createTime\s+=.*?;)`)
	//CopyrightStatPattern 无
	//VoiceListPattern 无
	AppMsgAlbumImgPattern         = regexp.MustCompile(`(?s)window.(appmsgalbuminfo\s+=\s+.*?];)`)
	PicturePageInfoListImgPattern = regexp.MustCompile(`(?s)window.(picture_page_info_list\s+=\s+.*?;)`)
)

var (
	BizVideoPattern       = regexp.MustCompile(`(?s)window.(biz\s+=\s+.*?;)`)
	SnVideoPattern        = regexp.MustCompile(`(?s)window.(sn\s+=\s+.*?;)`)
	WxidVideoPattern      = regexp.MustCompile(`(?s)user_name\s+=\s+.*?;`) // ---》注意
	AppMsgIdVideoPattern  = regexp.MustCompile(`(?s)window.(appmsgid\s+=\s+.*?;)`)
	MidVideoPattern       = regexp.MustCompile(`(?s)window.(mid\s+=\s+.*?;)`)
	ReqIdVideoPattern     = regexp.MustCompile(`(?s)req_id\s+=\s+.*?;`) // ---》注意
	AlbumIdVideoPattern   = regexp.MustCompile(`(?s)(album_id:.*?),`)
	AuthorVideoPattern    = regexp.MustCompile(`(?s)nick_name\s+=\s+.*?;`)
	AuthorIdVideoPattern  = regexp.MustCompile(`(?s)author_id\s+=\s+.*?;`)
	SessionIdVideoPattern = regexp.MustCompile(`(?s)window.(sessionid\s+=\s+.*?;)`)
	//ArticleTypePattern 无
	CommentIdVideoPattern     = regexp.MustCompile(`(?s)comment_id\s+=\s+.*?;`)
	CreateTimeVideoPattern    = regexp.MustCompile(`(?s)(var\s+createTime\s+=.*?;)`)
	CopyrightStatVideoPattern = regexp.MustCompile(`(?s)(ori_status:.*?),`)
	//VoiceListPattern 无
	MpVideoTransInfoVideoPattern = regexp.MustCompile(`(?s)window.(__mpVideoTransInfo\s+=\s+.*?];)`)
)
