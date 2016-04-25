//
//  self.swift
//  PetGag
//
//  Created by Rajdeep Mann on 3/29/16.
//  Copyright © 2016 PetGag. All rights reserved.
//

import UIKit


class CommentObject: NSObject{
    var userName: String!
    var comment: String!
}

class TrendingObject : NSObject{
    var imageId: String = "";
    var imageURLAsString: String = "";
    var numberOfUpVotes: Int = 0;
    var numberOfDownVotes: Int = 0;
    var comments: Array<CommentObject> = [];
    var image: UIImage!;
    
    
    override init() {
        
    }
    
    
    
    init(jsonInfo: NSDictionary) {
        
        self.imageURLAsString = (jsonInfo["ImageURL"]  as? String)!;
        self.numberOfUpVotes = (jsonInfo["UpVote"] as? Int)!;
        self.numberOfDownVotes = (jsonInfo["DownVote"] as? Int)!;
        self.imageId = (jsonInfo["_id"] as? String)!;
        
        let commentsData = jsonInfo["CommentList"] as? NSArray
        
        for comment in commentsData! {
            let commentObject = CommentObject.init()
            commentObject.userName = comment["UserCName"] as? String
            commentObject.comment = comment["Comment"] as? String
            self.comments.append(commentObject);
        }
    }
}

class PetGagAPI: NSObject,CLUploaderDelegate {

    
    var activityIndicator : UIActivityIndicatorView!;
    var baseView : UIView!;
    
    static let sharedInstane = PetGagAPI()
    
    override init() {
        super.init()
        let rect = CGRectMake(0,0,50,50);
        let window = UIApplication.sharedApplication().keyWindow! as UIWindow

        baseView = UIView.init(frame: (UIApplication.sharedApplication().keyWindow?.bounds)!);
        baseView.backgroundColor = UIColor.lightGrayColor();
        
        
        activityIndicator = UIActivityIndicatorView(frame: rect)
        activityIndicator.center = baseView.center;
        baseView.addSubview(activityIndicator)
        window.addSubview(baseView);
        
        activityIndicator.hidesWhenStopped = true;
        activityIndicator.stopAnimating();
        baseView.hidden = true;

    }
    
     func showActivityIndicator(){
        dispatch_async(dispatch_get_main_queue(),{
            
            let window = UIApplication.sharedApplication().keyWindow! as UIWindow
            self.baseView.removeFromSuperview()
            window.addSubview(self.baseView);

            self.activityIndicator.startAnimating();
            self.baseView.hidden = false;
            
        })

        
    }
    
     func hideActivityIndicator(){
        dispatch_async(dispatch_get_main_queue(),{
            
            self.baseView.removeFromSuperview()

            self.activityIndicator.stopAnimating();
            self.baseView.hidden = true;
            
        })
    }
    
    
    func fetchTrendingItems(completionHandler: (Array<TrendingObject>?) -> Void) ->Void{
        self.showActivityIndicator()
        
        let url = "http://localhost:8080/fetchAllPosts"
        
        let serverUrl = NSURL(string: url);
        let request = NSMutableURLRequest(URL:serverUrl!);
        // Excute HTTP Request
        let task = NSURLSession.sharedSession().dataTaskWithRequest(request) {
            data, response, error in
            
            // Check for error
            if error != nil
            {
                print("error=\(error)")
                return
            }
            
            // Print out response string
            let responseString = NSString(data: data!, encoding: NSUTF8StringEncoding)
            //                print("responseString = \(responseString)")
            
            // Convert server json response to NSDictionary
            do {
                if let jsonArray = try NSJSONSerialization.JSONObjectWithData(data!, options: []) as? NSArray {
                    
                    print("responseString = \(responseString)")
                   
                    var tempTrendingObjectsArray = [] as Array<TrendingObject>;
                    
                        for trendingObjectJSON in jsonArray {
                            let trendingObject = TrendingObject.init(jsonInfo: trendingObjectJSON as! NSDictionary)
                            tempTrendingObjectsArray.append(trendingObject);
                        }
                    self.hideActivityIndicator()
                    completionHandler(tempTrendingObjectsArray)
                }
            } catch let error as NSError {
                print(error.localizedDescription)
            }
            
        }
        
        task.resume()
        
        
    }
    
    func uploadImage(image: UIImage, onCompletion: (status: Bool, url: String?) -> Void) {
        
        
        let cloudinary_url = "cloudinary://491496947935867:okoNpZdxHxpMpE4nt1Em8pGXBmc@dsproject"
        
        let clouder = CLCloudinary(url:cloudinary_url)
        let forUpload = UIImageJPEGRepresentation(image,0.5)
        let uploader:CLUploader = CLUploader(clouder, delegate: self)
        
        uploader.upload(forUpload, options: nil,
                        withCompletion: { (dataDictionary: [NSObject: AnyObject]!, errorResult:String!, code:Int, context: AnyObject!) -> Void in
                            
                            print("response %@",dataDictionary)
                            let url = dataDictionary["url"] as! String

                            if code < 400 {
                                onCompletion(status: true, url:url)
                            }
                            else {
                                onCompletion(status: false, url: nil)
                            }
            },
                        andProgress: { (bytesWritten:Int, totalBytesWritten:Int, totalBytesExpectedToWrite:Int, context:AnyObject!) -> Void in
                            print("Upload progress: \((totalBytesWritten * 100)/totalBytesExpectedToWrite) %");
            }
        )
    }
    
        
    func postGag(gagInfo: NSMutableDictionary,completionHandler: (Bool) -> Void) -> Void {
        self.showActivityIndicator()


        let url = NSURL(string: "http://localhost:8080/addPost")!
        let request = NSMutableURLRequest(URL: url)
        request.HTTPMethod = "POST"
        
        let image = gagInfo["ImageFile"] as! UIImage;

        
        
        
        //resize image
        
        // post image on cloudinary and fetch url
        uploadImage(image, onCompletion: { (status, url) -> Void in
            
        let postData = ["UserName":NSUserDefaults.standardUserDefaults().valueForKey("fbusername")!,"ImageURL":url!] as NSDictionary
            
            
            
            request.setValue("application/json; charset=utf-8", forHTTPHeaderField: "Content-Type")
            
            do {
                
                let jsonData = try NSJSONSerialization.dataWithJSONObject(postData, options: [])
                
                request.HTTPBody = jsonData;
                
                
                
                // here "jsonData" is the dictionary encoded in JSON data
            } catch let error as NSError {
                print(error)
            }
            
            let task = NSURLSession.sharedSession().dataTaskWithRequest(request){ data,response,error in
                
                
                self.hideActivityIndicator()

                if error != nil{
                    //                print(error.localizedDescription)
                    return
                }
                
                
            }
            
            task.resume()
        
        });

    }
    
    
    func postComment(postId: String, userName: String, comment: String, completionHandler: (Bool) -> Void) -> Void {
        
        self.showActivityIndicator()

        let url = NSURL(string: "http://localhost:8080/addComment")!
        let request = NSMutableURLRequest(URL: url)
        request.HTTPMethod = "POST"
        
        let postData = ["UName":userName, "ImageId": postId, "Comment": comment];
        
        
        request.setValue("application/json; charset=utf-8", forHTTPHeaderField: "Content-Type")
            
            do {
                
                let jsonData = try NSJSONSerialization.dataWithJSONObject(postData, options: [])
                
                request.HTTPBody = jsonData;
                
                
                
                // here "jsonData" is the dictionary encoded in JSON data
            } catch let error as NSError {
                print(error)
            }
            
            let task = NSURLSession.sharedSession().dataTaskWithRequest(request){ data,response,error in
                
                
                
                self.hideActivityIndicator()

                if error != nil{
                    //                print(error.localizedDescription)
                    return
                }
                
                
            }
            
            task.resume()
        
    }
    
    
    func upVote(postId: String, completionHandler: (Bool) -> Void) -> Void {
        
        self.showActivityIndicator()

        
//        let delay = 4.5 * Double(NSEC_PER_SEC)
//        let time = dispatch_time(DISPATCH_TIME_NOW, Int64(delay))
//        dispatch_after(time, dispatch_get_main_queue()) {


        
        
        let url = NSURL(string: "http://localhost:8080/upVote")!
        let request = NSMutableURLRequest(URL: url)
        request.HTTPMethod = "POST"
        
        let postData = ["ImageId": postId];
        
        
        request.setValue("application/json; charset=utf-8", forHTTPHeaderField: "Content-Type")
        
        do {
            
            let jsonData = try NSJSONSerialization.dataWithJSONObject(postData, options: [])
            
            request.HTTPBody = jsonData;
            
            
            
            // here "jsonData" is the dictionary encoded in JSON data
        } catch let error as NSError {
            print(error)
        }
        
        let task = NSURLSession.sharedSession().dataTaskWithRequest(request){ data,response,error in

            self.hideActivityIndicator()

            completionHandler(error == nil);
            
        }
        
        task.resume()
//        }
    }
    
    func downVote(postId: String, completionHandler: (Bool) -> Void) -> Void {
        
        self.showActivityIndicator()

        
        let url = NSURL(string: "http://localhost:8080/downVote")!
        let request = NSMutableURLRequest(URL: url)
        request.HTTPMethod = "POST"
        
        let postData = ["ImageId": postId];
        
        
        request.setValue("application/json; charset=utf-8", forHTTPHeaderField: "Content-Type")
        
        do {
            
            let jsonData = try NSJSONSerialization.dataWithJSONObject(postData, options: [])
            
            request.HTTPBody = jsonData;

            // here "jsonData" is the dictionary encoded in JSON data
        } catch let error as NSError {
            print(error)
        }
        
        let task = NSURLSession.sharedSession().dataTaskWithRequest(request){ data,response,error in
            
            self.hideActivityIndicator()

            completionHandler(error == nil);
            
        }
        
        task.resume()
        
    }
    
}
