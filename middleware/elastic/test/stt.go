package test

type Admin struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	ParentId int    `json:"parent_id"`
	RootId   int    `json:"root_id"`
}

type AggeResultTopHits struct {
	DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int `json:"sum_other_doc_count"`
	Buckets                 []struct {
		Key      int `json:"key"`
		DocCount int `json:"doc_count"`
		TopRes   struct {
			Hits struct {
				Total struct {
					Value    int    `json:"value"`
					Relation string `json:"relation"`
				} `json:"total"`
				MaxScore float32 `json:"max_score"`
				Hits     []struct {
					Index  string  `json:"_index"`
					Type   string  `json:"_type"`
					Id     string  `json:"_id"`
					Score  float32 `json:"_score"`
					Source struct {
						ParentId int `json:"parent_id"`
					} `json:"_source"`
					Sort []int `json:"sort"`
				} `json:"hits"`
			} `json:"hits"`
		} `json:"top_res"`
	} `json:"buckets"`
}
type TopHits struct {
	TopTime AggeResultTopHits `json:"top_time"`
}
type AggeOrder struct {
	ParentId struct {
		DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int `json:"sum_other_doc_count"`
		Buckets                 []struct {
			Key        string `json:"key"`
			DocCount   int    `json:"doc_count"`
			CreateDate struct {
				Value int `json:"value"`
			} `json:"create_date"`
		} `json:"buckets"`
	} `json:"brand_id"`
}
