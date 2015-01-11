package main

import (
  "fmt"
  "strings"
  "net/http"
  "strconv"
  "github.com/PuerkitoBio/goquery"
  "encoding/json"
  //"os"
)

type FindResponse struct{
  Description     []MasterDescription
  Title string
  Number string
  Abstract string
  LegalEvents string
  Referenced string
  Citations string
  Classification string
  Images []string
  IndClaims string
  TotalClaims string
  Claims string

  Table string
  PublishedAs string
  Application bool
}

type para struct{
  Text string
  Id string
}

type MasterDescription struct{
  Title string
  TitleSlug string
  Paragraphs []para

}

func ExampleScrape(w http.ResponseWriter, r *http.Request){
  response := FindResponse{}
  args := r.FormValue("number")
  doc, _ := goquery.NewDocument("http://google.com/patents/" + args)
  if doc.Find("heading").Length() > 0 {
    response.Description = make([]MasterDescription, doc.Find("heading").Length())
    doc.Find("heading").Each(func(i int, f *goquery.Selection) {
      paragraphs := make([]para, f.Siblings().Length())


      f.Siblings().Each(func(j int, x *goquery.Selection) {
        if x.Is("p") {
          html, _ := x.Html()
          quick_id := fmt.Sprintf("pm-%v",j)
          small_paragraph := para{
            Text: strings.Replace(html,"\"","'",-1),
            Id: quick_id,
          }
          
          paragraphs = append(paragraphs, small_paragraph)
        }
      })
      if paragraphs == nil {
        paragraphs = make([]para, 0)
      }

      descrip := MasterDescription{
        Title: strings.TrimSpace(f.Text()),
        TitleSlug: strings.Join(strings.Split(strings.ToLower(strings.TrimSpace(f.Text())), " "), "_"),
        Paragraphs: paragraphs,

      }
      
      response.Description[i] = descrip
    })

  } else{
    paragraphs := make([]para, 1)
    response.Description = make([]MasterDescription, 1)
    this_text, _ := doc.Find(".patent-description-section").Html()
    quick_id := "old"

    small_paragraph := para{
      Text: strings.Replace(this_text,"\"","'",-1),
      Id:  quick_id,
    }
    paragraphs = append(paragraphs, small_paragraph)
    descrip := MasterDescription{
      Title: "Description",
      TitleSlug: "description",
      Paragraphs: paragraphs,
    }

    response.Description[0] = descrip
  }

  //Title and number
  response.Title = strings.TrimSpace(doc.Find(".patent-title").Text())
  response.Number = strings.TrimSpace(doc.Find(".patent-number").Text())
  response.Abstract, _ = doc.Find("abstract").Html()
  response.Abstract = strings.Replace(response.Abstract,"\"","'",-1)
  response.LegalEvents, _ = doc.Find("#legal-events~.patent-data-table").Html()
  response.LegalEvents = strings.Replace(response.LegalEvents,"\"","'",-1)

  if doc.Length() > 0 && doc.Find("#forward-citations").Length() > 0 {
    response.Referenced, _ = doc.Find("#forward-citations~.patent-data-table").Html()
    referenced, _ := doc.Find("#forward-citations~.patent-section-footer").Html()
    response.Referenced = strings.Replace(response.Referenced,"\"","'",-1) + strings.Replace(referenced,"\"","'",-1)
  } else {
    response.Referenced = ""
  }
  if doc.Length() > 0 && doc.Find("#backward-citations").Length() > 0 {
    response.Citations, _ = doc.Find("#backward-citations~.patent-data-table").Html()
    citations, _ := doc.Find("#backward-citations~.patent-section-footer").Html()
    response.Citations = strings.Replace(response.Citations,"\"","'",-1) + strings.Replace(citations,"\"","'",-1)
  } else {
    response.Citations = ""
  }
  if doc.Length() > 0 && doc.Find("#classifications").Length() > 0 {
    response.Classification, _ = doc.Find("#classifications~table.patent-data-table").Html()
    response.Classification = strings.Replace(response.Classification,"\"","'",-1)
  } else {
    response.Classification = ""
  }

  img_len := doc.Find(".patent-thumbnail-image").Length()
  response.Images = make([]string, doc.Find(".patent-thumbnail-image").Length())
  doc.Find(".patent-thumbnail-image").Each(func(i int, f *goquery.Selection) {
    src, _ := f.Attr("src")
    t := strings.Split(src,"thumbnails/")
    if i != (img_len-1) {
      if len(t) > 0 {
        response.Images = append(response.Images,(t[0]+t[1]))
      } else {
        response.Images = append(response.Images,(t[0]))
      }
    } else {
      if len(t) > 0 {
        response.Images = append(response.Images,(t[0]+t[1]))
      } else {
        response.Images = append(response.Images,(t[0]))
      }
    }
  })

  if doc.Length() > 0 && doc.Find(".patent-claims-section").Length() > 0 {
    response.Claims, _ = doc.Find(".patent-claims-section").Find(".claims").Html()
    response.Claims = strings.Replace(response.Claims,"\"","'",-1)
    response.IndClaims = strconv.Itoa(doc.Find(".claim .claim").Length())
    response.TotalClaims = strconv.Itoa(doc.Find(".claim-dependent").Length() + doc.Find(".claim .claim").Length())
  } else {
    response.Claims = ""
    response.IndClaims = "0"
    response.TotalClaims = "0"
  }
  published := strings.Split(doc.Find(".patent-bibdata-list-row").First().Find("td").Last().Find(".patent-bibdata-value-list").Text(), ", ")
  response.PublishedAs = "'" + strings.Join(published, "', '") + "'"
  if len(published) > 2 {
    small_table, _:= doc.Find(".patent-bibdata-list-row").First().Find("td").Last().Find(".patent-bibdata-value-list").Html()
    small_split := strings.Split(small_table, "<a ")
    if len(small_split) > 1 {
      larger_split := strings.Split(small_split[1], "</a>")[0]
      new_html := "<a " + larger_split + "</a>, <a href='' class='hideablePublished' data-list=\"" + strings.Join(published, ", ") + "\">" + strconv.Itoa(len(published)) + " More &#187;</a>"
      doc.Find(".patent-bibdata tr:not(.patent-bibdata-group-spacer):not(.alternate-patent-number):not(.patent-internal-links)").Find("td span.patent-bibdata-value-list").First().ReplaceWithHtml(new_html)
    }
  }

  doc.Find(".patent-bibdata tr:not(.patent-bibdata-group-spacer):not(.alternate-patent-number):not(.patent-internal-links)").Each(func(i int, s *goquery.Selection) {
    table,_ :=  s.Html()
    response.Table = response.Table + "<tr>" + table + "</tr>"
  })

  response.Table = strings.Replace(response.Table,"\"","'",-1)

  if strings.Index(doc.Find(".patent-bibdata tr:not(.patent-bibdata-group-spacer):not(.alternate-patent-number):not(.patent-internal-links)").Find("td.single-patent-bibdata").Text(),"Application") > -1 {
    response.Application = true
  } else {
    response.Application = false
  }

  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write(js)
}






func main(){
  http.HandleFunc("/", ExampleScrape)
  //err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
  err := http.ListenAndServe(":3000", nil)
  if err != nil {
    panic(err)
  }
}