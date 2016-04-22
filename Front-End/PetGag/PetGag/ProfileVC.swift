//
//  ProfileVC.swift
//  PetGag
//
//  Created by Rajdeep Mann on 3/31/16.
//  Copyright © 2016 PetGag. All rights reserved.
//

import UIKit

class ProfileVC: UIViewController, UITextFieldDelegate {

    var trendingObject : TrendingObject;
    let petGagAPI = PetGagAPI.sharedInstane;

    init(trendingObject:TrendingObject){
    
        self.trendingObject = trendingObject;
        super.init(nibName: nil, bundle: nil)
    }
    
    required init?(coder aDecoder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    override func loadView() {
        super.loadView();
        self.view.backgroundColor = UIColor.whiteColor()
        
        
        NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(ProfileVC.keyboardWillShow(_:)), name: UIKeyboardWillShowNotification, object: nil)
        NSNotificationCenter.defaultCenter().addObserver(self, selector: #selector(ProfileVC.keyboardWillHide(_:)), name: UIKeyboardWillHideNotification, object: nil)
        
        let imageName = "placeholder.png"
        
        var rect = CGRect(
            origin: CGPoint(x: 10, y: 74),
            size: CGSize(
                width: 355,
                height: 200
            )
        )
        
        let imageView = UIImageView(frame: rect) as UIImageView;
        imageView.userInteractionEnabled = true
        imageView.image = UIImage.init(named: imageName)
        imageView.frame = rect;
        
        imageView.downloadedFrom(link: trendingObject.imageURLAsString, contentMode: UIViewContentMode.ScaleToFill,trendingObject: trendingObject)
        
        self.view.addSubview(imageView)
     
        var yIndex = 0;
        
        for  commentObject in trendingObject.comments {
            
            var rect = CGRect(
                origin: CGPoint(x: 10, y: 300 + yIndex),
                size: CGSize(
                    width: 80,
                    height: 20
                )
            )
            
            let userNameLabel = UILabel.init(frame: rect)
            userNameLabel.font = UIFont.boldSystemFontOfSize(12);
            userNameLabel.text = commentObject.userName;
            self.view.addSubview(userNameLabel)
            
            rect.origin.x = 80;
            rect.size.width = 250;
            
            let commentLabel = UILabel.init(frame: rect)
            commentLabel.text = commentObject.comment;
            self.view.addSubview(commentLabel)
            
            yIndex += 25;
        }
        
        rect.origin.x = 0;
        rect.origin.y = self.view.bounds.size.height - 85;
        rect.size.width = 375;
        rect.size.height = 40;
        
        
        let commentField = UITextField.init(frame: rect)
        commentField.borderStyle = UITextBorderStyle.Bezel;
        commentField.placeholder = "Write a comment..."
        commentField.delegate = self;
        self.view.addSubview(commentField);
        
    }
    
    func textFieldShouldReturn(textField: UITextField) -> Bool{
        if textField.text != "" {
            self.view.endEditing(true);
            petGagAPI.postComment(self.trendingObject.imageId, userName: NSUserDefaults.standardUserDefaults().valueForKey("fbusername") as! String, comment: textField.text!, completionHandler: { (status) in
                
                
                
            });
            
        }
        return false;
    }
    
    func keyboardWillShow(notification: NSNotification) {
        
        if let keyboardSize = (notification.userInfo?[UIKeyboardFrameBeginUserInfoKey] as? NSValue)?.CGRectValue() {
            self.view.frame.origin.y -= keyboardSize.height
        }
        
    }
    
    func keyboardWillHide(notification: NSNotification) {
        if let keyboardSize = (notification.userInfo?[UIKeyboardFrameBeginUserInfoKey] as? NSValue)?.CGRectValue() {
            self.view.frame.origin.y += keyboardSize.height
        }
    }

    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
        // Dispose of any resources that can be recreated.
    }
    

    /*
    // MARK: - Navigation

    // In a storyboard-based application, you will often want to do a little preparation before navigation
    override func prepareForSegue(segue: UIStoryboardSegue, sender: AnyObject?) {
        // Get the new view controller using segue.destinationViewController.
        // Pass the selected object to the new view controller.
    }
    */

}
