//
//  HomeTableViewController.swift
//  PetGag
//
//  Created by Rajdeep Mann on 3/29/16.
//  Copyright © 2016 PetGag. All rights reserved.
//

import UIKit

import ObjectiveC

var ActionBlockKey: UInt8 = 0

// a type for our action block closure
typealias BlockButtonActionBlock = (sender: UIButton) -> Void

class ActionBlockWrapper : NSObject {
    var block : BlockButtonActionBlock
    init(block: BlockButtonActionBlock) {
        self.block = block
    }
}

extension UIButton {
    func block_setAction(block: BlockButtonActionBlock) {
        objc_setAssociatedObject(self, &ActionBlockKey, ActionBlockWrapper(block: block), objc_AssociationPolicy.OBJC_ASSOCIATION_RETAIN_NONATOMIC)
        addTarget(self, action: #selector(UIButton.block_handleAction(_:)), forControlEvents: .TouchUpInside)
    }
    
    func block_handleAction(sender: UIButton) {
        let wrapper = objc_getAssociatedObject(self, &ActionBlockKey) as! ActionBlockWrapper
        wrapper.block(sender: sender)
    }
}

extension UIImageView {
   
    func downloadedFrom(link link:String, contentMode mode: UIViewContentMode, trendingObject: TrendingObject) {
        guard
            let url = NSURL(string: link)
            else {return}
        contentMode = mode
        if trendingObject.image != nil {
            self.image = trendingObject.image;
        }else{
            
        
        NSURLSession.sharedSession().dataTaskWithURL(url, completionHandler: { (data, response, error) -> Void in
            guard
                
                let httpURLResponse = response as? NSHTTPURLResponse where httpURLResponse.statusCode == 200,
                let mimeType = response?.MIMEType where mimeType.hasPrefix("image"),
                let data = data where error == nil,
                let image = UIImage(data: data)
                
                else {
                    return
            }
            dispatch_async(dispatch_get_main_queue()) { () -> Void in
                self.image = image
                trendingObject.image = image;

            }
        }).resume()
           
        
        }
    }
}

class HomeTableViewController: UITableViewController, FBSDKLoginButtonDelegate {

    @IBOutlet weak var loginButton: FBSDKLoginButton!
    var trendingItems : Array<TrendingObject> = [];
    var petGagAPI: PetGagAPI!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        // Do any additional setup after loading the view, typically from a nib.

        self.petGagAPI = PetGagAPI.sharedInstane;
        
        self.title = "PetGag";

        let rect = CGRect(
            origin: CGPoint(x: 0, y: 10),
            size: CGSize(
                width: 375,
                height: 40
            )
        )
        
        self.tableView.tableHeaderView = UISearchBar(frame: rect)
        self.tableView.registerClass(UITableViewCell.self, forCellReuseIdentifier: "cell");


        if (FBSDKAccessToken.currentAccessToken() != nil)
        {
            //            let loginManager = FBSDKLoginManager()
            //            loginManager.logOut() // this is an instance function
            
            self.storeFBName();
            
            petGagAPI.fetchTrendingItems( { (trendingObjects) in
                self.trendingItems = trendingObjects!;
                
                dispatch_async(dispatch_get_main_queue(),{
                    
                    self.tableView.reloadData();
                    
                })

                
                
            });
            
            // User is already logged in, do work such as go to next view controller.
        }
        else
        {
            let loginView : FBSDKLoginButton = FBSDKLoginButton()
            self.view.addSubview(loginView)
            loginView.center = self.view.center
            loginView.readPermissions = ["public_profile", "email", "user_friends"]
            loginView.delegate = self
        }
    }

    func loginButton(loginButton: FBSDKLoginButton!, didCompleteWithResult result: FBSDKLoginManagerLoginResult!, error: NSError!) {
       
        print("User Logged In")
        
        if ((error) != nil)
        {
            // Process error
        }
        else if result.isCancelled {
            // Handle cancellations
        }
        else {
            loginButton.hidden = true
            petGagAPI.fetchTrendingItems( { (trendingObjects) in
                self.trendingItems = trendingObjects!;
                
                dispatch_async(dispatch_get_main_queue(),{
                    
                    self.tableView.reloadData();
                    
                })
                
                
                
            });
        }
    }
    
    
    
    func storeFBName() ->  Void {
        
        var fbUsername : String?
        
        let graphRequest : FBSDKGraphRequest = FBSDKGraphRequest(graphPath: "me", parameters: nil)
        graphRequest.startWithCompletionHandler({ (connection, result, error) -> Void in
            
            if ((error) != nil)
            {
                // Process error
                print("Error: \(error)")
            }
            else
            {
                
                fbUsername = result.valueForKey("name") as? String ?? ""
                NSUserDefaults.standardUserDefaults().setValue(fbUsername, forKey: "fbusername");
                
            }

        })
    }
    
    /*<class_name>.getFbUsername { (username) -> () in
    self.usernameLabel.text = username
    }*/
    
    func loginButtonDidLogOut(loginButton: FBSDKLoginButton!) {
        print("User Logged Out")
    }
    

    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
        // Dispose of any resources that can be recreated.
    }

    // MARK: - Table view data source

    override func numberOfSectionsInTableView(tableView: UITableView) -> Int {
        // #warning Incomplete implementation, return the number of sections
        return 1
    }

    override func tableView(tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        // #warning Incomplete implementation, return the number of rows
        return self.trendingItems.count;
    }

    
    override func tableView(tableView: UITableView, cellForRowAtIndexPath indexPath: NSIndexPath) -> UITableViewCell {

        let cell = tableView.dequeueReusableCellWithIdentifier("cell", forIndexPath: indexPath)
        cell.selectionStyle = UITableViewCellSelectionStyle.None;
        
        let imageName = "placeholder.png"
       
        let rect = CGRect(
            origin: CGPoint(x: 10, y: 10),
            size: CGSize(
                width: 355,
                height: 140
            )
        )
        
        let imageView = UIImageView(frame: rect) as UIImageView;
        imageView.userInteractionEnabled = true
        imageView.image = UIImage.init(named: imageName)
        imageView.frame = rect;
        
        imageView.downloadedFrom(link: trendingItems[indexPath.row].imageURLAsString, contentMode: UIViewContentMode.ScaleToFill,trendingObject: trendingItems[indexPath.row])
        
        cell.addSubview(imageView);
        
        let pointsLabel = UILabel.init(frame: CGRectMake(265,160,100,30));
        pointsLabel.textAlignment = NSTextAlignment.Right;
        cell.addSubview(pointsLabel);

        let points = trendingItems[indexPath.row].numberOfUpVotes - trendingItems[indexPath.row].numberOfDownVotes as Int;
        
        pointsLabel.text = NSString.init(format: "%d points", points) as String;

        
        let upVoteButton   = UIButton(type: UIButtonType.System) as UIButton
        upVoteButton.frame = CGRectMake(10, 160, 30, 30)
        
        let upVoteButtonImage = "thumb_up.png"
        
        upVoteButton.setImage(UIImage.init(named: upVoteButtonImage), forState: UIControlState.Normal)
        upVoteButton.tintColor = UIColor.darkGrayColor();
        
        cell.addSubview(upVoteButton)
        weak var weakSelf = self;
        
        upVoteButton.block_setAction { (sender) in
            weakSelf!.petGagAPI.upVote(weakSelf!.trendingItems[indexPath.row].imageId, completionHandler: { (status) in
                
                
                if status {

                    weakSelf!.trendingItems[indexPath.row].numberOfUpVotes = weakSelf!.trendingItems[indexPath.row].numberOfUpVotes + 1;
                    
                    let points = weakSelf!.trendingItems[indexPath.row].numberOfUpVotes - weakSelf!.trendingItems[indexPath.row].numberOfDownVotes as Int;
                    
                    dispatch_async(dispatch_get_main_queue(),{
                        
                        pointsLabel.text = NSString.init(format: "%d points", points) as String;
                        
                    })
                }
            })
        }
    
        
        let downVoteButton   = UIButton(type: UIButtonType.System) as UIButton
        downVoteButton.frame = CGRectMake(50, 160, 30, 30)
        
        let downVoteButtonImage = "thumb_down.png"
        
        downVoteButton.setImage(UIImage.init(named: downVoteButtonImage), forState: UIControlState.Normal)
        downVoteButton.tintColor = UIColor.darkGrayColor();
        cell.addSubview(downVoteButton)
        
        downVoteButton.block_setAction { (sender) in
            weakSelf!.petGagAPI.downVote(weakSelf!.trendingItems[indexPath.row].imageId, completionHandler: { (status) in
                
                
                if status {
                    weakSelf!.trendingItems[indexPath.row].numberOfDownVotes = weakSelf!.trendingItems[indexPath.row].numberOfDownVotes + 1;
                    
                    let points = weakSelf!.trendingItems[indexPath.row].numberOfUpVotes - weakSelf!.trendingItems[indexPath.row].numberOfDownVotes as Int;
                    
                    dispatch_async(dispatch_get_main_queue(),{
                        
                        pointsLabel.text = NSString.init(format: "%d points", points) as String;
                        
                    })
                }
                
            })
        }
        
        
        let commentsButton   = UIButton(type: UIButtonType.System) as UIButton
        commentsButton.frame = CGRectMake(93, 160, 30, 30)
        
        let commentsButtonImage = "comments.png"
        
        commentsButton.setImage(UIImage.init(named: commentsButtonImage), forState: UIControlState.Normal)
        commentsButton.tintColor = UIColor.darkGrayColor();
        cell.addSubview(commentsButton)
        
        
        let line = UIView(frame: CGRect(x: 0, y: 199, width: 375, height: 1));
        line.backgroundColor = UIColor.lightGrayColor();
        
        
        cell.addSubview(line)
        
        return cell
    }
    
    
    override func tableView(tableView: UITableView, didSelectRowAtIndexPath indexPath: NSIndexPath){
        
        let trendingObject = self.trendingItems[indexPath.row] as TrendingObject;
        let profileVC = ProfileVC.init(trendingObject: trendingObject);
        self.navigationController?.pushViewController(profileVC, animated: true)
    }
    
   
    @IBAction func showNewPostView(sender: UIBarButtonItem) {

        let storyboard = UIStoryboard(name: "Main", bundle: nil)
        let vc = storyboard.instantiateViewControllerWithIdentifier("NewPostVC") 
        self.navigationController?.pushViewController(vc, animated: true)
        
    }

    
    
}
