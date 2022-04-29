package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/rendicott/uggly"
	"github.com/rendicott/uggo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"strings"
	"io/ioutil"
	"encoding/json"
	"net/http"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	port       = flag.Int("port", 10000, "The server port")
)

var loremIpsum string = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

`

type Elixer struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Effect string `json:"effect"`
	SideEffects string `json:"sideEffects"`
	Characteristics string `json:"characteristics"`
	Time string `json"time"`
	Manufacturer string `json:"manufacturer"`
}

type Wizard struct {
	Elixers []Elixer `json:"elixirs"`
	Id string `json:"id"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
}

func getWizards() (wizards []Wizard, err error) {
	response, err := http.Get("https://wizard-world-api.herokuapp.com/Wizards")
	if err != nil {
		return wizards, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return wizards, err
	}
	err = json.Unmarshal(responseData, &wizards)
	if err != nil {
		return wizards, err
	}
	return wizards, err
}

var strokeMap = []string{"1","2","3","4","5","6","7","8","9",
	"a","b","c","d","e","f","g","h","i","j","k","l","m",
	"n","o","p","q","r","s","t","u","v","w","x","y","z"}

func addWizardBox(inPage *pb.PageResponse) (*pb.PageResponse, error) {
	var err error
	// first grab content so we can size boxes right
	wizards, err := getWizards()
	if err != nil { return inPage, err }
	inPage.DivBoxes.Boxes = append(inPage.DivBoxes.Boxes, &pb.DivBox{
		Name:     "wizards",
		Border:   false,
		FillChar: convertStringCharRune(""),
		StartX:   3,
		StartY:   3,
		Width:    int32(30),
		Height:   int32(len(wizards)+3),
		FillSt: &pb.Style{
			Fg:   "grey",
			Bg:   "black",
			Attr: "4",
		},
	})
	contentString := fmt.Sprintf("  WIZARDS  \n", )
	lmap := make(map[string]string)
	for i, wiz := range(wizards) {
		name := fmt.Sprintf("%s %s", wiz.FirstName, wiz.LastName)
		lmap[name] = strokeMap[i]
	}
	for name, stroke := range lmap {
		inPage.KeyStrokes = append(inPage.KeyStrokes, &pb.KeyStroke{
			KeyStroke: stroke,
			Action: &pb.KeyStroke_Link{
				Link: &pb.Link{
					PageName: name,
					Server: "localhost",
					Port: "8888",
			}}})
		contentString += fmt.Sprintf("(%s) %s\n", stroke, name)
	}
	inPage.Elements.TextBlobs = append(inPage.Elements.TextBlobs, &pb.TextBlob{
		Content: contentString,
		Wrap:    true,
		Style: &pb.Style{
			Fg:   "white",
			Bg:   "black",
			Attr: "4",
		},
		DivNames: []string{"wizards"},
	})
	return inPage, err
}

func okay(preq *pb.PageRequest) (*pb.PageResponse,error) {
	var err error
	height := int(preq.ClientHeight)
	width := int(preq.ClientWidth)
	log.Printf("func ok height %d, width %d", height, width)
	links := []*uggo.PageLink{
		&uggo.PageLink{
			Page: "one",
			KeyStroke: "1",
		},
		&uggo.PageLink{
			Page: "two",
			KeyStroke: "2",
		},
		&uggo.PageLink{
			Page: "three",
			KeyStroke: "3",
		},
		&uggo.PageLink{
			Page: "four",
			KeyStroke: "4",
		},
	}
	return uggo.PageTopMenuFullWidthContent(
		width, height, links,preq.Name,okContent[preq.Name]), err
}

func wizards(preq *pb.PageRequest) (presp *pb.PageResponse, err error) {
	height := int(preq.ClientHeight)
	width := int(preq.ClientWidth)
	localPage := pb.PageResponse{
		Name: preq.Name,
		DivBoxes: &pb.DivBoxes{},
		Elements: &pb.Elements{},
	}
	localPage.DivBoxes.Boxes = append(localPage.DivBoxes.Boxes, &pb.DivBox{
		Name:     "dynamic-main",
		Border:   false,
		FillChar: convertStringCharRune("|"),
		StartX:   0,
		StartY:   0,
		Width:    int32(width),
		Height:   int32(height),
		FillSt: &pb.Style{
			Fg:   "grey",
			Bg:   "black",
			Attr: "4",
		},
	})
	finalPage, err := addWizardBox(&localPage)
	return finalPage, err
}

func flipFlopColor(num int) (string) {
	if num%2 == 0 {
		return "darkslategrey"
	} else {
		return "springgreen"
	}
}

func shelp(fg, bg string) *pb.Style {
	return &pb.Style{
		Fg:   fg,
		Bg:   bg,
		Attr: "4",
	}
}

func formSubmit(ctx context.Context, preq *pb.PageRequest) (presp *pb.PageResponse, err error) {
	height := int(preq.ClientHeight)
	width := int(preq.ClientWidth)
	localPage := pb.PageResponse{
		Name: "formResponse",
		DivBoxes: &pb.DivBoxes{},
		Elements: &pb.Elements{},
	}
	localPage.DivBoxes.Boxes = append(localPage.DivBoxes.Boxes, &pb.DivBox{
		Name:     "formR",
		Border:   true,
		BorderW:  int32(1),
		BorderChar: convertStringCharRune("^"),
		FillChar: convertStringCharRune(""),
		StartX:   5,
		StartY:   5,
		Width:    int32(width-10),
		Height:   int32(height-15),
		BorderSt: shelp("orange", "black"),
		FillSt: shelp("white","black"),
	})
	var name string
	var age string
	for _, fd := range preq.FormData {
		for _, td := range fd.TextBoxData {
			if td.Name == "name" {
				name = td.Contents
			}
			if td.Name == "age" {
				age = td.Contents
			}
		}
	}
	msg := fmt.Sprintf("Hi, %s, I see that you're %s.", name, age)
	localPage.Elements.TextBlobs = append(localPage.Elements.TextBlobs, &pb.TextBlob{
		Content: msg,
		Wrap:    true,
		Style: &pb.Style{
			Fg:   "white",
			Bg:   "black",
			Attr: "4",
		},
		DivNames: []string{"formR"},
	})
	localPage.SetCookies = append(localPage.SetCookies, &pb.Cookie{
		Key: "name",
		Value: name,
	})
	localPage.SetCookies = append(localPage.SetCookies, &pb.Cookie{
		Key: "age",
		Value: age,
	})
	return &localPage, err
}

func form(ctx context.Context, preq *pb.PageRequest) (presp *pb.PageResponse, err error) {
	height := int(preq.ClientHeight)
	width := int(preq.ClientWidth)
	localPage := pb.PageResponse{
		Name: "form",
		DivBoxes: &pb.DivBoxes{},
		Elements: &pb.Elements{},
	}
	welcomeMessage := "hi, I don't think we've met before"
	var name string
	var age string
	for _, cookie := range preq.SendCookies {
		log.Printf("got cookie from client key = '%s', val = '%s'", cookie.Key, cookie.Value)
		if cookie.Key == "name" {
			name = cookie.Value
		}
		if cookie.Key == "age" {
			age = cookie.Value
		}
	}
	if name != "" && age != "" {
		welcomeMessage = fmt.Sprintf("Welcome back %s, are you still %s", name, age)
	}
	localPage.DivBoxes.Boxes = append(localPage.DivBoxes.Boxes, &pb.DivBox{
		Name:     "formDiv",
		Border:   true,
		BorderW:  int32(1),
		BorderChar: convertStringCharRune("^"),
		FillChar: convertStringCharRune(""),
		StartX:   5,
		StartY:   5,
		Width:    int32(width-10),
		Height:   int32(height-15),
		BorderSt: shelp("orange", "black"),
		FillSt: shelp("white","black"),
	})
	localPage.Elements.Forms = append(localPage.Elements.Forms, &pb.Form{
		Name: "test",
		DivName: "formDiv",
		SubmitLink: &pb.Link{
			PageName: "formSubmit",
		},
		TextBoxes: []*pb.TextBox{
			&pb.TextBox{
				Name: "name",
				TabOrder: 2,
				DefaultValue: "<your name here>",
				Description: "Name: ",
				PositionX: 25,
				PositionY: 10,
				Height: 1,
				Width: 30,
				StyleCursor: shelp("black", "gray"),
				StyleFill: shelp("black", "blue"),
				StyleText: shelp("white", "blue"),
				StyleDescription: shelp("red", "black"),
				ShowDescription: true,
			},
			&pb.TextBox{
				Name: "age",
				TabOrder: 4,
				DefaultValue: "<your age here>",
				Description: "Age: ",
				PositionX: 25,
				PositionY: 12,
				Height: 1,
				Width: 30,
				StyleCursor: shelp("black", "gray"),
				StyleFill: shelp("black", "blue"),
				StyleText: shelp("white", "blue"),
				StyleDescription: shelp("red", "black"),
				ShowDescription: true,
			},
		},
	})
	localPage.KeyStrokes = append(localPage.KeyStrokes, &pb.KeyStroke{
		KeyStroke: "j",
		Action: &pb.KeyStroke_FormActivation{
			FormActivation: &pb.FormActivation{
				FormName: "test",
	}}})
	localPage.Elements.TextBlobs = append(localPage.Elements.TextBlobs, &pb.TextBlob{
		Content: welcomeMessage,
		Wrap:    true,
		Style: &pb.Style{
			Fg:   "white",
			Bg:   "black",
			Attr: "4",
		},
		DivNames: []string{"formDiv"},
	})
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("form got incoming metadata: %v", md)
	} else {
		log.Print("form no metadata received")
	}
	return &localPage, err
}

func wacky(preq *pb.PageRequest) (presp *pb.PageResponse, err error) {
	height := int(preq.ClientHeight)
	width := int(preq.ClientWidth)
	log.Printf("got new client width, height: %d, %d\n", width, height)
	cellWidth := width / 7
	cellHeight := height / 6
	localPage := pb.PageResponse{
		Name: preq.Name,
		DivBoxes: &pb.DivBoxes{},
		Elements: &pb.Elements{},
	}
	for j:=0; j<=cellHeight; j++ {
		for i:=0; i<=cellWidth; i++ {
			localPage.DivBoxes.Boxes = append(localPage.DivBoxes.Boxes, &pb.DivBox{
				Name:     fmt.Sprintf("cell-%d", i),
				Border:   false,
				FillChar: convertStringCharRune(""),
				StartX:   int32(i*cellWidth),
				StartY:   int32(j*cellHeight),
				Width:    int32(cellWidth),
				Height:   int32(cellHeight),
				FillSt: &pb.Style{
					Fg:   "grey",
					Bg:   flipFlopColor(i+j),
					Attr: "4",
				},
			})
		}
	}
	boxWidth := width - width/4
	boxHeight := height - height/5
	localPage.DivBoxes.Boxes = append(localPage.DivBoxes.Boxes, &pb.DivBox{
		Name:     "content",
		Border:   true,
		BorderW:  int32(1),
		BorderChar: convertStringCharRune("^"),
		FillChar: convertStringCharRune(""),
		StartX:   int32(width/2 - boxWidth/2),
		StartY:   int32(height/2 - boxHeight/2),
		Width:    int32(boxWidth),
		Height:   int32(boxHeight),
		BorderSt: &pb.Style{
			Fg:   "darkolivegreen",
			Bg:   "lightgreen",
			Attr: "4",
		},
		FillSt: &pb.Style{
			Fg:   "grey",
			Bg:   "black",
			Attr: "4",
		},
	})
	localPage.Elements.TextBlobs = append(localPage.Elements.TextBlobs, &pb.TextBlob{
		Content: strings.Repeat(loremIpsum, 3),
		Wrap:    true,
		Style: &pb.Style{
			Fg:   "white",
			Bg:   "black",
			Attr: "4",
		},
		DivNames: []string{"content"},
	})
	return &localPage, err
}

/* GetPage implements the Page Service's GetPage method as required in the protobuf definition.

It is the primary listening method for the server. It accepts a PageRequest and then attempts to build
a PageResponse which the client will process and display on the client's pcreen. 
*/
func (s pageServer) GetPage(ctx context.Context, preq *pb.PageRequest) (presp *pb.PageResponse, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("got incoming metadata: %v", md)
	} else {
		log.Print("no metadata received")
	}
	if preq.Name == "home" {
		return wacky(preq)
	} else if preq.Name == "form" {
		return form(ctx, preq)
	} else if preq.Name == "formSubmit" {
		return formSubmit(ctx, preq)
	} else if preq.Name == "one" || preq.Name == "two" || preq.Name == "three" || preq.Name == "four" {
		return okay(preq)
	} else {
		return wizards(preq)
	}
}

/* newPageServer takes the loaded pageconfig YAML and converts it to the structs
required so that the GetPage method can adequately respond with a PageResponse.
*/
func newPageServer() *pageServer {
	pServer := &pageServer{}
	return pServer
}


/* pageServer is a struct from which to attach the required methods for the Page Service
as defined in the protobuf definition
*/
type pageServer struct {
	pb.UnimplementedPageServer
	//pages []*pageServerPage
}

// convertStringCharRune takes a string and converts it to a rune slice
// then grabs the rune at index 0 in the slice so that it can return
// an int32 to satisfy the Uggly protobuf struct for border and fill chars
// and such. If the input string is less than zero length then it will just
// rune out a space char and return that int32.
func convertStringCharRune(s string) int32 {
	if len(s) == 0 {
		s = " "
	}
	runes := []rune(s)
	return runes[0]
}

/* feedServer is a struct from which to attach the required methods for the Feed Service
as defined in the protobuf definition
*/
type feedServer struct {
	pb.UnimplementedFeedServer
	pages []*pb.PageListing
}

/* newFeedServer generates a feed of pages this server wants to expose in an
index to a client that requests it
*/
func newFeedServer() *feedServer {
	fServer := &feedServer{}
	// ./server.go:82:17: first argument to append must be slice; have *uggly.Pages
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "wizards",
	})
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "home",
	})
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "form",
	})
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "one",
	})
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "two",
	})
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "three",
	})
	fServer.pages = append(fServer.pages, &pb.PageListing{
		Name: "four",
	})
	return fServer
}

/* GetFeed implements the Feed Service's GetFeed method as required in the protobuf definition.

It is the primary listening method for the server. It accepts a FeedRequest and then attempts to build
a FeedResponse which the client will process. 
*/
func (f feedServer) GetFeed(ctx context.Context, freq *pb.FeedRequest) (fresp *pb.FeedResponse, err error) {
	fresp = &pb.FeedResponse{}
	fresp.Pages = f.pages
	return fresp, err
}


func main() {
	flag.Parse()
	genOkContent()
	//lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	f := newFeedServer()
	pb.RegisterFeedServer(grpcServer, *f)
	s := newPageServer()
	pb.RegisterPageServer(grpcServer, *s)
	log.Println("Server listening")
	grpcServer.Serve(lis)
}

var okContent map[string]string

func genOkContent() {
	okContent = make(map[string]string)
	okContent["one"] = `If you're already familiar with networking then you can jump to the [technical explanation](./3pocalypse-primer) this part or re-read for a refresher.

## Synopsis (Non-Technical Analogy)
As we all are probably aware the Internet is comprised of a finite number of IP addresses. These addresses are like the mailing address of your home or apartment. They provide a place where anyone in the world can send information to you. 

There are only about 3.7 billion possible publicly routable IP addresses on the internet. If the same was true for physical addresses there would only be about 3.7 billion addresses for any home or business in the world. This obviously doesn't scale very well for internet things or people for that matter so a different system had to be designed. 

Let's take for example an office building with thousands of people in it at thousands of desks on hundreds of floors. The address of this building is "123 Broad Street". In order for Jane on the 55th floor sitting in zone 6 to recieve her letter she must be able to tell someone how to send her a letter. She calls up her friend Joe and tells him that in order to send her a letter he must use two envelopes. On one envelope he writes down "Jane" and then he puts his letter inside that envelope. Then he takes the other envelope and he writes "123 Broad Street" on it and he puts the other envelope inside that one. So now we have an envelope inside of an envelope. The address on the outer envelope is the internet routable address.

The mailman knows how to get the outer envelope to the building but that's all they know how to do. Once the mail reaches the building someone in the mail room opens the envelope and sees that it's addressed to "Jane" and looks inside their building directory and sees that Jane is on the 55th Floor Zone 6 and knows how to get that letter to Jane. The same is true in the reverse. 

On the inner envelope that Jane received there was a return address. Jane knows she needs to send her response to Joe at 546 Willow Street but she doesn't exactly where in the world that is but she's pretty sure it's not in the building. Someone told her that the mailroom knows how to send stuff outside the building. So, Jane puts her reply letter in an envelope and on that envelope she writes "Joe". She puts that envelope in another envelope and on that envelope she writes "546 Willow Street" and sends it down to the mail room. 

Every day the BGP comes by and tells the mail room all the street names in the city and the next mail office they should send letters for those street names. This is kind of like the BGP route sharing that happens on the internet. 

The mail room receives Jane's letter and knows immediately that the street address is not inside the building since the BGP told them so. They send it out to a regional mail office where they hope it can be sorted out. The regional mail office receives it and looks in their directory and sees that Broad street is close by so they send out a regional mail carrier to Joes building and the process is reversed. 

The above example is highly simplified but you can think of the building's public addresses (e.g., 123 Broad Street) as the public routable internet IP addresses, the floor and zone numbers (e.g., "55th floor Zone 6") as the private routable IP addresses, the mail room is the Gateway, and their directories are their route tables.

#### A Step Further
Let's say that for the General Electric company instead of using a system like "55th floor, zone 6" the building had mimicked that street naming convention like the rest of the world has done. The building itself is 123 Willow Lane and they call everything in the building "Willow Lane" and everyone on every floor gets their own address on Willow Lane. E.g., Jane on the 55th floor zone 6 is 5506 Willow Lane. This is quite convenient since you can use the same naming convention to have people send mail from outside the building as you can inside the building. You can even use some special addresses and share those with people outside of the building and they still know how to get letters to you. 

When you want to send a letter to someone in the same building you can just address them directly by only using one envelope and sending it straight to them at an address like 2299 Willow Lane without having to go through the mail room.

So, in this scenario when Joe sends Jane a letter he just puts "123 Willow Lane" on the outer envelope and still writes "Jane" on the inner envelope. When the mail room opens the outer envelope they look in the directory and they see that Jane resides at 5506 Willow Lane so they know that's on the 55th Floor Zone 6 so they can get her the letter. This is great and everyone at GE is very happy with this system.

#### The Problem
Now let's say that the world has completely run out of space on the ground to build buildings and no street has any space left. The only option is for people build their buildings taller and use the same street addresses. GE decides that they could make a lot of money by selling their spot on Willow Lane and moving their facilities inside someone else's sky scraper so they build on top of the building at 123 Baker Street. However, they don't want to go through the headache of re-labeling everyone's desk so everyone keeps their same desk addresses. Instead, they just tell the mail room that whenever they open a letter and the inner envelope is addressed to someone at Willow Lane that they still look for that person within the building. 

Now that GE have vacated the real Willow Lane a new company has moved into the building on 123 Willow Lane and has started operating. They're so happy that GE decided to vacate the property so Jim decides to send a thank you letter to Jane. On the outer envelope Jim writes "123 Baker Street" and on the inner letter he writes "Jane". When the GE mailroom receives this letter they send it to Jane at her 5506 Willow Lane desk location within the Baker Street building. She's very happy with the letter and writes a reply. She puts her reply letter in an envelope labeled "Jim" and puts it in an envelope labeled "123 Willow Lane" and sends it to the mail room. 

That same day the BGP came by and told the mail room that "all letters going to Willow Lane go outside" which is normal behavior for the mail room at Baker Street.  

When the mail room receives the letter they see the address as Willow Lane so they are very confused. They err on the side of comfort and tradition and naturally look within their internal directory and try it send it back upstairs to someone named Jim on the 44th floor. Jim sends it back saying he doesn't know what this is all about. The mail room tries a few more times then throws it in the trash.

Jim never receives a reply to his thank you letter and is forced to assume that Jane is a rude person. 

#### Solutions
There are a few possible solutions to GE's Willow Lane predicament above. We'll go through a few of them and try to explain the challenges:

* Everyone could renumber their desks within the building
   * This is a possibility. However, what do you do about the hundreds of people who have their coworker's internal desk addresses memorized? You'd have to go around and tell everyone the new desk number and hope that none of the thousands of daily internal memos get lost.
   * This will take a lot of time to teach everyone their new desk number. Sometimes there's not even a real person at the desk.
* You could tell the mail room that some letters destined for "Willow Lane" are actually outside the building.
   * What do you do when two people have the same address? When a letter comes in for 7865 Willow Lane is that the outside one or the inside one?

At this point I think I've run this analogy into the ground but I think you get the idea.`
	okContent["two"] = `This article is an overview of problems stemming from the sale of the '3.0.0.0/8' address space and the phenomenon known within GE as "3pocalypse". It is intended to be a very high level overview for a wider audience.

If you're not familiar with networking it may be helpful to skip the technical synopsis and read the [non-technical analogy](./3pocalypse-analogy) and then come back to get an idea of what's going on.

**Note:** Please be aware that the issue has been largely mitigated by the changes the proxy team has implemented on the pZEN appliances (access points to PITC) in the various sites (Cincinnati, Tokyo, Amsterdam, etc.)
They are now able to distinguish where to route based on where the request came from, rather than purely on a static route to 3.0.0.0/8, therefore this is no longer an issue for any person or server using the PAC file or explicitly pointing to one of the patched endpoints
Please refer to [Yammer post](https://www.yammer.com/ge.com/#/Threads/show?threadId=332707483680768) and [PITC info](https://internet.ge.com/docs/integration-guide/) for more information

## Synopsis (Technical)
GE once owned the entire '3.0.0.0/8' block of publicly routable IP address space. These days it is incredibly rare for a non-ISP or Cloud provider to own such a massive number of public IPs (roughly 16.7 million addresses) due to the fact that there is a finite number of publicly routable addresses (about 3.7 billion total) and they are currently very expensive to buy. Most companies don't need this many addresses and can get away with using private address space with a few hundred public IP's using Network Address Translation (NAT) for public facing services.

In early 2018 it was [announced](https://news.ycombinator.com/item?id=18407173) that Amazon had purchased the '3.0.0.0/8' address space from GE to use for their AWS cloud services. 

GE had already beeen in the process of moving all devices with a 3.x.x.x IP address into the '10.0.0.0/8' address space for a few years but GE has not been able to move all services and there are still hundreds of applications depending on internal 3.x.x.x. addresses. Unfortunately, Amazon has already started allowing 3.x.x.x internet routable IP addresses to be allocated to customer devices within their AWS cloud. This means that a startup company could create the next Instagram or Facebook and start serving traffic on a 3.x.x.x. IP address if their stuff is hosted in AWS. 

This puts a lot of things at GE in a predicament in a few different ways:
* When traffic originating from an Amazon 3.x.x.x. IP address enters a GE DMZ it gives its return address as 3.x.x.x. and the GE router thinks that's still internal so it can't send the return traffic back.
* Things on GE's edge (internet facing services) sometimes blindly trust anything originating from a 3.x.x.x. IP address since that used to be the universal indicator that it was safe internal traffic. 
* Things within GE sending traffic destined for services within AWS that have a 3.x.x.x. address will probably never make it out of the GE network and therefore are inaccessible. 
* Workloads running inside AWS DirectConnect enabled VPC's will have a tough time deciding on how to talk to 3.x.x.x. IP addresses as more and more AWS services and customers begin using the IP's. 

There are a lot of potential solutions to this but many of them are very disruptive and expensive. 
* What if we just change the IP's from 3.x.x.x. to 10.x.x.x. ?
    * If you just re-IP everything in the company it would cause hundreds of outages since a lot of things don't use DNS and have 3.x.x.x. as their target destination for services. 
    * How do you choose which 10.x.x.x. address to give an existing 3.x.x.x device? Chances are that it's not a straight search and replace since the 10.x.x.x. addresses have already been in use. 
    * How do you parse through thousands of firewall rules and update them programatically? Some firewall rules are a range of IP's and some are individual. Who's to say what will break if a rule that mentions a range of 3.x.x.x. IP addresses is removed or modified.
* What if we just refuse to talk to internet 3.x.x.x. stuff?
    * This would cut out a lot of things hosted in AWS including our own stuff!
    * This will only work for a little while until more and more external services and vendors start getting 3.x.x.x. IP addresses. Think about our parts suppliers, emergency alert systems, chat providers, etc. We'd have to refuse talking to them and sometimes that's not an option.

Many different networking teams at GE have been discussing the best way to approach this problem over the past year and unfortunately there are no one-size-fits-all solutions. All solutions will have to be executed carefully and could be very disruptive to existing services. 

For now, the best way is to tackle each problem as it comes up. This will probably involve projects to isolate internally facing devices from externally facing devices and make sure their route tables and security rules are set up to talk to their intended destinations. 

In the beginning these problems will manifest themselves as sporadic, intermittent issues. Sometimes traffic will enter a system and not be able to return. Sometimes AWS services will receive a 3.x.x.x. IP address, not be able to talk to GE, then their address will change back to a 54.x.x.x. (another common AWS IP range) and start working again miraculously.

The worst-case scenario would be a GE box on the internet edge with a public IP that blindly trusts any device with a 3.x.x.x. IP and then an attacker somehow gaining access to that resource. 

Either way this is going to be something that GE will be solving on a case-by-case basis over time. If you have specific questions about your application please see the "Where to Get Help" section at the bottom of this article. 



## Known Issues
In this section we'll list some known issues that have occurred as a result of this phenomenon.

* Internet Lambdas`
}
