/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package login

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/krmdv/cli/api"
	"github.com/krmdv/cli/config"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdLogin creates a login command
func NewCmdLogin() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "login <token>",
		Short: "Login to Karma",
		Long:  `Login to Karma`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return loginRun(args[0])
		},
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringP("token", "t", "", "your karma api token")
	cmd.Flags().StringP("org", "o", "", "active github organization")
	cmd.Flags().StringP("slack", "s", "", "your slack notifications webhook URL")

	return cmd
}

func loginRun(token string) error {
	type userRes struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}

	var user userRes

	resp, _, errs := gorequest.New().
		Get(config.Host()+"/users/me").
		Set("Authorization", fmt.Sprintf("token %s", token)).
		EndStruct(&user)

	if len(errs) != 0 {
		return errs[0]
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	if !success {
		return api.HandleHTTPError(resp)
	}

	viper.Set("user.id", user.ID)
	viper.Set("user.name", user.Name)
	viper.Set("token", token)
	viper.Set("users", "")
	viper.Set("team", "")
	viper.Set("feats", "")

	home, _ := homedir.Dir()

	viper.WriteConfigAs(home + "/.karma.yaml")

	fmt.Println(`
		                                                                                
                    ((,,                                     ,,,,               
                   (###,,,                                 .*,,/(*              
                  ((/###/,,                               **,,((#/              
                 ,((*/###(,,,                           /**,/(#//(/             
                 (#***/###(*,,,                       */***(#(/,,#//            
                /(#****//##(/,,*                     //**/(#(/,,.,(/            
               #*#(****//(##(/***                  .///*/##(/*,...(//           
               ((##***///(*##(/***.      , *      /((///##(//*,...,//           
               ((##**////(( ##(//*//(///**(***/#*/##((/(#(///*,.../((           
               /(#/*/////((.###(((/**********,,,****/((##((//*,,.,((/           
                /((/////((##(/********,,,,,,,,,,,,,,,,,***/((/**,//#,           
                 *((((((#((/******,,,,,,,,,,,,,,,,,,,,,,,,**/(###(              
                   / ##((//*,..,,,,,,,,,,,,,,,,,,,,,,    ,,***///               
                   *##((,.           ,,,,,,,,,,,            ,**///,             
                  /(((*...,(&%%&/.    *,,,,,,,    .*&&%&#,.  .**///             
                 ,(((/..,/. %%&.%,,   ,,,,,,,,  .,#,#%%./.(. .,**///            
                .(((//,.,*.(&&@@(*,...**//*//*,..,,*&&@@( #..,,**////*          
               /((((//*,.   ...,,,,,*************,,,,,..   .,,****//(((*        
       ,*,,,   .*##((//***************/&&@&@&/*********,,,,*******/#(((         
    ,**,,,*//  ##(((#((///((////////////,(#*/////////****//////(######(         
   //*/(###/(*  ,##%%%########((((((((((((((/////////((////(##(((#####          
  /////((####*   ,/((########(((((///////**//****///(##%%##%######((            
 ./////((((((* ///((((((((((/**,,,,,,,,,,,,,,,,,,  .,,,///////*,,,,//           
 *////(((((##########((((***,,..     .,,,,,,,,        ,,,,,,***,,,,,,**         
 */((((((((#########(((****......                       .,,,,,****,,,,**        
  (((((((##########((/***.........                        ,,,,**//**,,,***      
  .(((((########%##(//*,,,..........                       .,,,**///******/     
    (((######%%%%##(/**,,,,.............               ......****/((//*****//   
     ,/((###%%%%##((/***,,,,,.................................**//(##(//**///(  
          /########(******,,,,,,..............................,//((###((/////(, 
              /####(//*******,,,,,,,,,,,...............,,,,,,,,/(((####(((((((( 
              /(####/////*********,,,,,,,,,,,,,,,,,,,,,,,,,,,**(((##(%###(((((( 
              *(####/////////*********************************/#####(%%######(. 
               /(####////////////////********************/////######/%%####(/   
                /((####/////////////////////////////////////#######((((,*,.     
                 /((((####(//////////////////////////////#########(*            
                  (((((((##########(////////////((###############(/             
                  ./(####################(((#####################(              
                   /(####%%%%%%%%%%%(.       ../(((##%%%%%#(((((#/              
                   /###############              ,(#%%%###(((###(               
                     %%%%%%%%%%%#                 #%%%%%%%%%%%%                 

    :...::...:::........::........:::......::::.......:::..:::::..::........::....:: 
    '##:::::'##:'########:'##::::::::'######:::'#######::'##::::'##:'########:'####:
    :##:'##: ##: ##.....:: ##:::::::'##... ##:'##.... ##: ###::'###: ##.....:: ####:
    :##: ##: ##: ##::::::: ##::::::: ##:::..:: ##:::: ##: ####'####: ##::::::: ####:
    :##: ##: ##: ######::: ##::::::: ##::::::: ##:::: ##: ## ### ##: ######:::: ##::
    :##: ##: ##: ##...:::: ##::::::: ##::::::: ##:::: ##: ##. #: ##: ##...:::::..:::
    :##: ##: ##: ##::::::: ##::::::: ##::: ##: ##:::: ##: ##:.:: ##: ##:::::::'####:
    . ###. ###:: ########: ########:. ######::. #######:: ##:::: ##: ########: ####:
    :...::...:::........::........:::......::::.......:::..:::::..::........::....::                                                                           
                                                                       
	
	`)

	color.Green(fmt.Sprintf("✅ Successfully logged in as %s.", user.Name))
	color.Yellow("👉 Please run 'karma config --org GITHUB_ORG' to setup your team and get started")

	return nil
}
