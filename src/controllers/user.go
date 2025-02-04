package controllers

import (
	"backend/src/db"
	"backend/src/models"
	"context"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UserService struct {
	DbAdapter *db.DbAdapter
}

func NewUserService(dbAdapter *db.DbAdapter) *UserService {
	return &UserService{DbAdapter: dbAdapter}
}
func (u UserService) RegisterParticipants(w http.ResponseWriter, r *http.Request) bool {
	// Parse the form data (max memory usage: 10MB for file uploads)
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return false
	}

	// Extract `participants` from form data
	participantsStr := r.FormValue("participants")
	transactionID := r.FormValue("transactionId")
	referralCode := r.FormValue("referralCode")

	// Debugging: Print raw form values
	log.Println("Participants (raw):", participantsStr)
	log.Println("Transaction ID:", transactionID)

	if participantsStr == "" || transactionID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return false
	}

	// Parse `participants` JSON string into a slice of maps
	var participantsData []map[string]interface{}
	if err := json.Unmarshal([]byte(participantsStr), &participantsData); err != nil {
		log.Println("JSON parsing error:", err)
		http.Error(w, "Invalid participants JSON", http.StatusBadRequest)
		return false
	}

	// Handle transaction image upload
	file, _, err := r.FormFile("transactionImage")
	if err != nil {
		http.Error(w, "Transaction screenshot is required", http.StatusBadRequest)
		return false
	}
	defer file.Close()

	ctx := context.Background()
	imageURL, is := u.FileUpload(ctx, file)
	if !is {
		http.Error(w, "Image upload failed", http.StatusInternalServerError)
		return false
	}

	// Insert participants into DB
	var participantIDs []int
	var participantNamesEmails []map[string]string
	for _, participantMap := range participantsData {
		name, _ := participantMap["name"].(string)
		email, _ := participantMap["email"].(string)
		phone, _ := participantMap["phone"].(string)
		collegeName, _ := participantMap["collegeName"].(string)
		yearOfStudy, _ := participantMap["yearOfStudy"].(float64) // JSON numbers default to float64
		dualBoot, _ := participantMap["dualBoot"].(bool)

		participant := models.Participant{
			Name:        name,
			Email:       email,
			Phone:       phone,
			CollegeName: collegeName,
			YearOfStudy: int(yearOfStudy),
			DualBoot:    dualBoot,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		pid, err := u.DbAdapter.CreateParticipant(ctx, participant)
		if err != nil {
			http.Error(w, "Error creating participant", http.StatusInternalServerError)
			return false
		}

		participantIDs = append(participantIDs, pid)
		participantNamesEmails = append(participantNamesEmails, map[string]string{
			"name": participant.Name, "email": participant.Email,
		})
	}

	// Create registration record
	registration := models.Registration{
		NumOfParticipants: len(participantIDs),
		Participants:      participantIDs,
		TotalAmount:       0,
		TransactionID:     transactionID,
		TransactionImage:  imageURL,
		ReferralCode:      referralCode,
	}

	if _, err := u.DbAdapter.CreateRegistration(ctx, registration); err != nil {
		http.Error(w, "Error creating registration", http.StatusInternalServerError)
		return false
	}

	// Send confirmation emails
	for _, participant := range participantNamesEmails {
		go u.SendEmail(models.Participant{Name: participant["name"], Email: participant["email"]},
			participantNamesEmails)
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))

	return true
}

func (u UserService) SendEmail(user models.Participant, participants []map[string]string) bool {
	from := os.Getenv("BACKEND_MAIL_USER")
	password := os.Getenv("BACKEND_MAIL_PASSWORD")
	host := os.Getenv("BACKEND_MAIL_HOST")
	to := user.Email

	log.Println(from, password, host, to)
	participantNames := ""
	for _, participant := range participants {
		participantNames += participant["name"] + "<br/>"
	}

	auth := smtp.PlainAuth("", from, password, host)

	emailTemplate := u.GetEmail(user.Name, participantNames)

	msg := []byte("From: " + from + "\r\n" + "To: " + to + "\r\n" +
		"Subject: Welcome to Metamorphosis\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		emailTemplate)

	err := smtp.SendMail(host+":587", auth, from, []string{to}, msg)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (u UserService) GetEmail(name string, participantNames string) string {
	return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link
      href="https://fonts.googleapis.com/css2?family=Poppins:ital,wght@0,400;0,500;0,600;1,400;1,500&display=swap"
      rel="stylesheet"
    />
    <title>METAMORPHOSIS 2K25</title>

    <!-- <title>Responsive GIF Display</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background-color: #f0f0f0;
            text-align: center;
        }
        .gif-container {
            max-width: 600;
            overflow: hidden;
        }
        img {
            width: 600;
            display: block;
        }
        
    </style> -->
  </head>

  <body style="font-family: 'Poppins', sans-serif">
    <div>
      <u></u>

      <div
        style="
          text-align: center;
          margin: 0;
          padding-top: 10px;
          padding-bottom: 10px;
          padding-left: 0;
          padding-right: 0;
          background-color: #f2f4f6;
          color: #000000;
        "
        align="center"
      >
        <div style="text-align: center">
          <table
            align="center"
            style="
              text-align: center;
              vertical-align: middle;
              width: 600px;
              max-width: 600px;
            "
            width="600"
          >
            <tbody>
              <tr>
                <td
                  style="width: 596px; vertical-align: middle"
                  width="596"
                ></td>
              </tr>
            </tbody>
          </table>
          
          <!-- <div class="gif-container"> -->
          <img
            style="text-align: center"
            alt="META 2K25 Banner"
            src="https://res.cloudinary.com/dfuwno067/image/upload/v1738444246/META_Banner_e0joky.png"
            width="600"
            class="CToWUd a6T"
            data-bit="iit"
            tabindex="0"
          />

          <div
            class="a6S"
            dir="ltr"
            style="opacity: 0.01; left: 552px; top: 501.5px"
          >
            <div
              id=":155"
              class="T-I J-J5-Ji aQv T-I-ax7 L3 a5q"
              role="button"
              tabindex="0"
              aria-label="Download attachment "
              jslog="91252; u014N:cOuCgd,Kr2w4b,xr6bB; 4:WyIjbXNnLWY6MTc2MjU0MTQxMTA0MjYyMTM2NyIsbnVsbCxbXV0."
              data-tooltip-class="a1V"
              data-tooltip="Download"
            >
              <div class="akn">
                <div class="aSK J-J5-Ji aYr"></div>
              </div>
            </div>
          </div>

          <table
            align="center"
            style="
              text-align: center;
              vertical-align: top;
              width: 600px;
              max-width: 600px;
              background-color: #ffffff;
            "
            width="600"
          >
            <tbody style="color: #343434">
              <tr>
                <td
                  style="
                    width: 596px;
                    vertical-align: top;
                    padding-left: 30px;
                    padding-right: 30px;
                    padding-top: 30px;
                    padding-bottom: 40px;
                  "
                  width="596"
                >
                  <h1
                    style="
                      font-size: 22px;
                      line-height: 34px;
                      font-family: 'Helvetica', Arial, sans-serif;
                      font-weight: 600;
                      text-decoration: none;
                      color: #000000;
                    "
                  >
                    Hola Tech Enthusiasts! 🐧
                  </h1>

                  <p
                    style="
                      line-height: 24px;
                      font-weight: 400;
                      text-decoration: none;
                    "
                  >
                    We are pleased to inform you that your registration for
                    <strong>MetaMorphosis 2K25</strong> was successful! 🎉<br /><br />
                    The event will be held on
                    <strong><em>15th & 16th of February, 2025</em></strong
                    >, focusing on Docker & Kubernetes.💜
                  </p>
                  <p>
                    <strong>Participant Name(s):</strong><br />
                    ` + participantNames + `
                  </p>
                  You will have access to all the sessions and activities we
                  have scheduled for the event as a registered participant.
                  <br />
                  <p>
                    Details of the event are as follows: <br />
                    <strong>Date:</strong> 15th & 16th of February, 2025 <br />
                    <strong>Time:</strong> 9:00 AM <br />
                    <strong>Venue:</strong>
                    Main & Mini CCF, WCE
                  </p>
                  Please do not hesitate to contact us if you have any queries
                  about the event. We will be happy to assist you in any way we
                  can.
                  <p></p>
                  <p>
                    <strong style="font-size: 17px">
                      MetaMorphosis 2K25 Website:</strong
                    >
                    <a
                      href="https://meta2k25.wcewlug.org/"
                      style="font-size: 17px"
                      >meta2k25.wcewlug.org</a
                    >
                    <br />
                    Do share this with your friends and join us for an exciting
                    journey!
                  </p>

                  <p>
                    <strong>
                      <i>We look forward to seeing you there!</i>
                    </strong>
                  </p>

                  <p>
                    Thanks and regards,<br />
                    Walchand Linux Users' Group
                  </p>
                </td>
              </tr>
            </tbody>
          </table>

        <table
          align="center"
          style="
            text-align: center;
            vertical-align: top;
            width: 600px;
            max-width: 600px;
            background-color: #ffffff;
          "
          width="600"
        >
          <tbody>
            <tr>
              <td
                style="
                  width: 600px;
                  vertical-align: top;
                  padding-left: 0;
                  padding-right: 0;
                "
              >
                <img
                  style="
                    text-align: center;
                    border-top-left-radius: 30px;
                    border-bottom-right-radius: 30px;
                    margin-bottom: 5px;
                  "
                  alt="Logo"
                  src="https://res.cloudinary.com/dduur8qoo/image/upload/v1689771850/wlug_white_logo_page-0001_u8efnh.jpg"
                  align="center"
                  width="200"
                  height="120"
                  class="CToWUd"
                  data-bit="iit"
                />
              </td>
            </tr>

            <tr style="margin-bottom: 30px" align="center">
              <td align="center">
              
                <a
                  href="https://linkedin.com/company/wlug-club"
                  target="_blank"
                  data-saferedirecturl="https://www.google.com/url?q=https://linkedin.com/company/wlug-club&amp;source=gmail&amp;ust=1680976985984000&amp;usg=AOvVaw0TDo2Akq1O-un9s_gRi70t"
                  style="margin: 0 10px"
                  ><img
                    src="https://res.cloudinary.com/dduur8qoo/image/upload/v1685247353/linkedin_mg2ujv.png"
                    class="CToWUd"
                    data-bit="iit"
                    height="30"
                    width="30"
                    style="border-radius: 5px"
                /></a>
                <a
                  href="http://discord.wcewlug.org/join"
                  target="_blank"
                  data-saferedirecturl="https://www.google.com/url?q=http://discord.wcewlug.org/join&amp;source=gmail&amp;ust=1680976985984000&amp;usg=AOvVaw3PNiAyDSeiO1V36KVKeLZl"
                  style="margin: 0 1px"
                  ><img
                    src="https://res.cloudinary.com/dduur8qoo/image/upload/v1689771996/unnamed_m7lgs0.png"
                    class="CToWUd"
                    data-bit="iit"
                    height="30"
                    width="30"
                    style="border-radius: 5px"
                /></a>
                <a
                  href="https://www.instagram.com/wcewlug/"
                  target="_blank"
                  data-saferedirecturl="https://www.google.com/url?q=https://www.instagram.com/wcewlug/&amp;source=gmail&amp;ust=1680976985984000&amp;usg=AOvVaw16ObtJOZ1hpw9644RZ4oMM"
                  style="margin: 0 12px"
                  ><img
                    src="https://res.cloudinary.com/dduur8qoo/image/upload/v1689773467/Instagram_vn7dni_kzulby.png"
                    class="CToWUd"
                    data-bit="iit"
                    height="30"
                    width="30"
                /></a>
                <a
                  href="https://twitter.com/wcewlug"
                  target="_blank"
                  data-saferedirecturl="https://www.google.com/url?q=https://twitter.com/wcewlug&amp;source=gmail&amp;ust=1680976985984000&amp;usg=AOvVaw1ypHRKREADjq_cn0IRD2po"
                  ><img
                    src="https://res.cloudinary.com/dfuwno067/image/upload/v1738444243/twitter_wxkrwu.png"
                    class="CToWUd"
                    data-bit="iit"
                    height="30"
                    width="30"
                    style="border-radius: 5px"
                /></a>
              </td>
            </tr>
          </tbody>
        </table>
          <div class="yj6qo"></div>
          <div class="adL"></div>
        </div>
        <div class="adL"></div>
      </div>
      <div class="adL"></div>
    </div>
  </body>
</html>`
}

func (u UserService) FileUpload(ctx context.Context, file multipart.File) (string, bool) {
	cld, _ := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_KEY"), os.Getenv("CLOUDINARY_SECRET"))

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "metamorphosis",
	})

	if err != nil {
		return "", false
	}

	return uploadResult.SecureURL, true

}
